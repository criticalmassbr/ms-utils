package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/approle"
)

type VaultConfig struct {
	RoleId    string
	SecretId  string
	Url       string
	MountPath string
	Cert      string
}

type IVaultService interface {
	NewVaultService(cfg *VaultConfig) IVaultService
	GetSecret(key VaultSecretKey, clientSlug string) (string, error)
	GetSecrets(clientSlug string, keys []string) (map[string]interface{}, error)
}

type VaultService struct {
	config *VaultConfig
	client *api.Client
}

type VaultSecretKey string

var Vault = VaultService{}

func (v *VaultService) NewVaultService(cfg *VaultConfig) IVaultService {
	service := &VaultService{
		config: cfg,
	}
	service.init()
	return service
}

func (s *VaultService) init() error {
	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(s.config.Cert)
	if err != nil {
		return fmt.Errorf("unable to read Vault certificate: %v", err)
	}
	certs.AppendCertsFromPEM(pemData)

	vaultConfig := &api.Config{
		Address: s.config.Url,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: certs,
				},
			},
		},
	}

	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize Vault client: %v", err)
	}
	s.client = client

	_, err = s.login()
	if err != nil {
		return fmt.Errorf("unable to login to Vault: %v", err)
	}
	go s.renewToken()
	return nil
}

func (s *VaultService) login() (*api.Secret, error) {
	appRoleAuth, err := s.newAppRoleAuth()
	if err != nil {
		return nil, err
	}

	authInfo, err := s.client.Auth().Login(context.Background(), appRoleAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AppRole auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return authInfo, nil
}

func (s *VaultService) newAppRoleAuth() (*auth.AppRoleAuth, error) {
	secretID := &auth.SecretID{FromString: s.config.SecretId}
	appRoleAuth, err := auth.NewAppRoleAuth(
		s.config.RoleId,
		secretID,
	)
	if err != nil {
		return nil, err
	}
	return appRoleAuth, nil
}

func (s *VaultService) renewToken() {
	for {
		_, span := Tracer.NewSpan(context.Background(), "vault", "Vault Token Renewal")

		authInfo, err := s.login()
		if err != nil {
			Tracer.AddSpanErrorAndFail(span, err, "unable to authenticate to Vault")
		}

		if err := s.manageTokenLifecycle(authInfo); err != nil {
			Tracer.AddSpanErrorAndFail(span, err, "unable to start managing token lifecycle")
		}

		span.End()
	}
}

func (s *VaultService) manageTokenLifecycle(token *api.Secret) error {
	if !token.Auth.Renewable {
		_, span := Tracer.NewSpan(context.Background(), "vault", "Vault Token Renewal")
		Tracer.AddSpanEvents(span, "Token not renewable", map[string]string{"message": "Token is not configured to be renewable. Re-attempting login."})
		span.End()
		return nil
	}

	watcher, err := s.client.NewLifetimeWatcher(&api.LifetimeWatcherInput{
		Secret:    token,
		Increment: 3600,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize new lifetime watcher for renewing auth token: %w", err)
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		_, span := Tracer.NewSpan(context.Background(), "vault", "Vault Token Renewal")

		select {
		case err := <-watcher.DoneCh():
			if err != nil {
				Tracer.AddSpanEvents(span, "Failed to renew token", map[string]string{"message": fmt.Sprintf("Failed to renew token: %v. Re-attempting login.", err)})
				span.End()
				return nil
			}
			Tracer.AddSpanEvents(span, "Failed to renew token", map[string]string{"message": "Token can no longer be renewed. Re-attempting login."})
			span.End()
			return nil

		case renewal := <-watcher.RenewCh():
			Tracer.AddSpanEvents(span, "Token renewed", map[string]string{"message": fmt.Sprintf("Successfully renewed: %#v", renewal)})
		}
		span.End()
	}
}

func (s *VaultService) GetSecret(key VaultSecretKey, clientSlug string) (string, error) {
	secret, err := s.client.KVv1(s.config.MountPath).Get(context.Background(), clientSlug)
	if err != nil {
		return "", fmt.Errorf("unable to read secret: %v", err)
	}

	value, ok := secret.Data["data"].(map[string]interface{})[string(key)].(string)
	if !ok {
		return "", fmt.Errorf("secret value type assertion failed")
	}

	return value, nil
}

func (s *VaultService) GetSecrets(clientSlug string, keys []string) (map[string]interface{}, error) {
	secrets, err := s.client.KVv1(s.config.MountPath).Get(context.Background(), clientSlug)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %v", err)
	}

	values, ok := secrets.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("secret value type assertion failed")
	}

	filteredSecrets := make(map[string]interface{})
	for _, key := range keys {
		if val, ok := values[key]; ok {
			filteredSecrets[key] = val
		}
	}

	return filteredSecrets, nil
}

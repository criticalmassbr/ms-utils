package vault

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"

	utils "github.com/criticalmassbr/ms-utils"
	"github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/approle"
)

type vaultRepository struct {
	config *VaultConfig
	client *api.Client
}

func NewVaultRepository(cfg *VaultConfig) (VaultRepository, error) {
	service := &vaultRepository{
		config: cfg,
	}
	err := service.init()
	return service, err
}

func (c *vaultRepository) init() error {
	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(c.config.Cert)
	if err != nil {
		return fmt.Errorf("unable to read Vault certificate: %v", err)
	}
	certs.AppendCertsFromPEM(pemData)

	vaultConfig := &api.Config{
		Address: c.config.Url,
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
	c.client = client

	_, err = c.login()
	if err != nil {
		return fmt.Errorf("unable to login to Vault: %v", err)
	}
	go c.renewToken()
	return nil
}

func (c *vaultRepository) login() (*api.Secret, error) {
	appRoleAuth, err := c.newAppRoleAuth()
	if err != nil {
		return nil, err
	}

	authInfo, err := c.client.Auth().Login(context.Background(), appRoleAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AppRole auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return authInfo, nil
}

func (c *vaultRepository) newAppRoleAuth() (*auth.AppRoleAuth, error) {
	secretID := &auth.SecretID{FromString: c.config.SecretId}
	appRoleAuth, err := auth.NewAppRoleAuth(
		c.config.RoleId,
		secretID,
	)
	if err != nil {
		return nil, err
	}
	return appRoleAuth, nil
}

func (c *vaultRepository) renewToken() {
	for {
		_, span := utils.Tracer.NewSpan(context.Background(), "vault", "Vault Token Renewal")

		authInfo, err := c.login()
		if err != nil {
			utils.Tracer.AddSpanErrorAndFail(span, err, "unable to authenticate to Vault")
		}

		if err := c.manageTokenLifecycle(authInfo); err != nil {
			utils.Tracer.AddSpanErrorAndFail(span, err, "unable to start managing token lifecycle")
		}

		span.End()
	}
}

func (c *vaultRepository) manageTokenLifecycle(token *api.Secret) error {
	if !token.Auth.Renewable {
		_, span := utils.Tracer.NewSpan(context.Background(), "vault", "Vault Token Renewal")
		utils.Tracer.AddSpanEvents(span, "Token not renewable", map[string]string{"message": "Token is not configured to be renewable. Re-attempting login."})
		span.End()
		return nil
	}

	watcher, err := c.client.NewLifetimeWatcher(&api.LifetimeWatcherInput{
		Secret:    token,
		Increment: 3600,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize new lifetime watcher for renewing auth token: %w", err)
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		_, span := utils.Tracer.NewSpan(context.Background(), "vault", "Vault Token Renewal")

		select {
		case err := <-watcher.DoneCh():
			if err != nil {
				utils.Tracer.AddSpanEvents(span, "Failed to renew token", map[string]string{"message": fmt.Sprintf("Failed to renew token: %v. Re-attempting login.", err)})
				span.End()
				return nil
			}
			utils.Tracer.AddSpanEvents(span, "Failed to renew token", map[string]string{"message": "Token can no longer be renewed. Re-attempting login."})
			span.End()
			return nil

		case renewal := <-watcher.RenewCh():
			utils.Tracer.AddSpanEvents(span, "Token renewed", map[string]string{"message": fmt.Sprintf("Successfully renewed: %#v", renewal)})
		}
		span.End()
	}
}

func (c *vaultRepository) GetSecrets(clientSlug string) (map[string]interface{}, error) {
	secret, err := c.client.KVv1(fmt.Sprintf("%s/data", c.config.MountPath)).Get(context.Background(), clientSlug)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %v", err)
	}

	secrets, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("secret value type assertion failed")
	}

	return secrets, nil
}

func (c *vaultRepository) List() ([]string, error) {
	secrets, err := c.client.Logical().List(fmt.Sprintf("%s/metadata", c.config.MountPath))
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %v", err)
	}

	if secrets.Data == nil {
		err := errors.New("unable to read secret")
		for _, warning := range secrets.Warnings {
			err = errors.Join(err, fmt.Errorf(warning))
		}
		return nil, err
	}

	values, ok := secrets.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("secret value type assertion failed")
	}

	keys := make([]string, 0)
	for _, value := range values {
		keys = append(keys, fmt.Sprintf("%v", value))
	}

	return keys, nil
}

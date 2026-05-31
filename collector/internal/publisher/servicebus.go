package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

type ServiceBusPublisher struct {
	sender *azservicebus.Sender
}

// New creates a publisher for the given Service Bus namespace and topic.
// It selects the credential based on AUTH_METHOD env var:
//
//	"cli"                → AzureCliCredential (local dev, requires az login)
//	"" / "managed_identity" → ManagedIdentityCredential (production default)
func New(namespace, topic string) (*ServiceBusPublisher, error) {
	cred, err := newCredential()
	if err != nil {
		return nil, fmt.Errorf("create credential: %w", err)
	}
	client, err := azservicebus.NewClient(namespace, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("create service bus client: %w", err)
	}
	sender, err := client.NewSender(topic, nil)
	if err != nil {
		return nil, fmt.Errorf("create sender for topic %q: %w", topic, err)
	}
	return &ServiceBusPublisher{sender: sender}, nil
}

func newCredential() (azcore.TokenCredential, error) {
	switch os.Getenv("AUTH_METHOD") {
	case "cli":
		return azidentity.NewAzureCLICredential(nil)
	default: // "managed_identity" or unset
		opts := &azidentity.ManagedIdentityCredentialOptions{}
		if id := os.Getenv("AZURE_CLIENT_ID"); id != "" {
			opts.ID = azidentity.ClientID(id)
		}
		return azidentity.NewManagedIdentityCredential(opts)
	}
}

func (p *ServiceBusPublisher) Send(ctx context.Context, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	msg := &azservicebus.Message{Body: body}
	if err := p.sender.SendMessage(ctx, msg, nil); err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

func (p *ServiceBusPublisher) Close(ctx context.Context) error {
	return p.sender.Close(ctx)
}

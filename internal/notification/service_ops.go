package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// SendDeploymentSuccess notifies a user that a deployment succeeded.
func (s *Service) SendDeploymentSuccess(ctx context.Context, userID uuid.UUID, appID, appName string, deployedAt int64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeDeploymentSuccess,
		Category: CategoryDeployment,
		Title:    fmt.Sprintf("Deployment Successful: %s", appName),
		Body:     fmt.Sprintf("Your deployment of %s was successful.", appName),
		Data: JSONMap{
			"app_id":      appID,
			"app_name":    appName,
			"deployed_at": deployedAt,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendDeploymentFailure notifies a user that a deployment failed.
func (s *Service) SendDeploymentFailure(ctx context.Context, userID uuid.UUID, appID, appName, errorMsg, logsURL string, failedAt int64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeDeploymentFailed,
		Category: CategoryDeployment,
		Title:    fmt.Sprintf("Deployment Failed: %s", appName),
		Body:     fmt.Sprintf("Your deployment of %s failed. %s", appName, errorMsg),
		Data: JSONMap{
			"app_id":        appID,
			"app_name":      appName,
			"error_message": errorMsg,
			"logs_url":      logsURL,
			"failed_at":     failedAt,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendFailoverTriggered notifies a user when failover is triggered for their function.
func (s *Service) SendFailoverTriggered(ctx context.Context, userID uuid.UUID, functionID, functionName, reason string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeFailoverTriggered,
		Category: CategoryFailover,
		Title:    fmt.Sprintf("Failover Triggered: %s", functionName),
		Body:     fmt.Sprintf("Failover was triggered for %s. Reason: %s.", functionName, reason),
		Data: JSONMap{
			"function_id":   functionID,
			"function_name": functionName,
			"reason":        reason,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendFailoverResolved notifies a user when failover resolves and normal operation resumes.
func (s *Service) SendFailoverResolved(ctx context.Context, userID uuid.UUID, functionID, functionName string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeFailoverResolved,
		Category: CategoryFailover,
		Title:    fmt.Sprintf("Failover Resolved: %s", functionName),
		Body:     fmt.Sprintf("Failover has resolved and normal operation has resumed for %s.", functionName),
		Data: JSONMap{
			"function_id":   functionID,
			"function_name": functionName,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendProviderOffline notifies a user when a provider they use goes offline.
func (s *Service) SendProviderOffline(ctx context.Context, userID uuid.UUID, providerID, providerName string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeProviderOffline,
		Category: CategoryProvider,
		Title:    fmt.Sprintf("Provider Offline: %s", providerName),
		Body:     fmt.Sprintf("Provider %s is now offline. Some operations may be affected.", providerName),
		Data: JSONMap{
			"provider_id":   providerID,
			"provider_name": providerName,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendProviderOnline notifies a user when a provider comes back online.
func (s *Service) SendProviderOnline(ctx context.Context, userID uuid.UUID, providerID, providerName string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeProviderOnline,
		Category: CategoryProvider,
		Title:    fmt.Sprintf("Provider Online: %s", providerName),
		Body:     fmt.Sprintf("Provider %s is now back online.", providerName),
		Data: JSONMap{
			"provider_id":   providerID,
			"provider_name": providerName,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendProviderDegraded notifies a user when a provider they use is degraded.
func (s *Service) SendProviderDegraded(ctx context.Context, userID uuid.UUID, providerID, providerName, reason string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeProviderDegraded,
		Category: CategoryProvider,
		Title:    fmt.Sprintf("Provider Degraded: %s", providerName),
		Body:     fmt.Sprintf("Provider %s is experiencing degraded performance. %s", providerName, reason),
		Data: JSONMap{
			"provider_id":   providerID,
			"provider_name": providerName,
			"reason":        reason,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

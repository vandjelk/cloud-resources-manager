package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsnfsvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
)

type NukeClient interface {
	awsnfsvolumebackupclient.Client

	// WebACL methods
	ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error)
	GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error)
	DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error
	ListTagsForWebACL(ctx context.Context, resourceArn string) ([]wafv2types.Tag, error)

	// Certificate methods
	ListCertificates(ctx context.Context) ([]acmtypes.CertificateSummary, error)
	ListCertificateTags(ctx context.Context, arn string) ([]acmtypes.Tag, error)
	DeleteCertificate(ctx context.Context, arn string) error
}

func NewClientProvider() awsclient.SkrClientProvider[NukeClient] {
	return func(ctx context.Context, account, region, key, secret, role string) (NukeClient, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}

		backupClient := backup.NewFromConfig(cfg)
		wafv2Client := wafv2.NewFromConfig(cfg)
		acmClient := acm.NewFromConfig(cfg)

		return &client{
			backupClient: awsnfsvolumebackupclient.NewClient(backupClient),
			wafv2Client:  wafv2Client,
			acmClient:    acmClient,
		}, nil
	}
}

type client struct {
	backupClient awsnfsvolumebackupclient.Client
	wafv2Client  *wafv2.Client
	acmClient    *acm.Client
}

// Embed backup client methods
func (c *client) IsNotFound(err error) bool {
	return c.backupClient.IsNotFound(err)
}

func (c *client) IsAlreadyExists(err error) bool {
	return c.backupClient.IsAlreadyExists(err)
}

func (c *client) ListTags(ctx context.Context, resourceArn string) (map[string]string, error) {
	return c.backupClient.ListTags(ctx, resourceArn)
}

func (c *client) ListBackupVaults(ctx context.Context) ([]backuptypes.BackupVaultListMember, error) {
	return c.backupClient.ListBackupVaults(ctx)
}

func (c *client) DescribeBackupVault(ctx context.Context, backupVaultName string) (*backup.DescribeBackupVaultOutput, error) {
	return c.backupClient.DescribeBackupVault(ctx, backupVaultName)
}

func (c *client) CreateBackupVault(ctx context.Context, name string, tags map[string]string) (string, error) {
	return c.backupClient.CreateBackupVault(ctx, name, tags)
}

func (c *client) DeleteBackupVault(ctx context.Context, name string) error {
	return c.backupClient.DeleteBackupVault(ctx, name)
}

func (c *client) StartBackupJob(ctx context.Context, params *awsnfsvolumebackupclient.StartBackupJobInput) (*backup.StartBackupJobOutput, error) {
	return c.backupClient.StartBackupJob(ctx, params)
}

func (c *client) DescribeBackupJob(ctx context.Context, backupJobId string) (*backup.DescribeBackupJobOutput, error) {
	return c.backupClient.DescribeBackupJob(ctx, backupJobId)
}

func (c *client) ListRecoveryPointsForVault(ctx context.Context, accountId, backupVaultName string) ([]backuptypes.RecoveryPointByBackupVault, error) {
	return c.backupClient.ListRecoveryPointsForVault(ctx, accountId, backupVaultName)
}

func (c *client) DescribeRecoveryPoint(ctx context.Context, accountId, backupVaultName, recoveryPointArn string) (*backup.DescribeRecoveryPointOutput, error) {
	return c.backupClient.DescribeRecoveryPoint(ctx, accountId, backupVaultName, recoveryPointArn)
}

func (c *client) DeleteRecoveryPoint(ctx context.Context, backupVaultName, recoveryPointArn string) (*backup.DeleteRecoveryPointOutput, error) {
	return c.backupClient.DeleteRecoveryPoint(ctx, backupVaultName, recoveryPointArn)
}

func (c *client) StartCopyJob(ctx context.Context, params *awsnfsvolumebackupclient.StartCopyJobInput) (*backup.StartCopyJobOutput, error) {
	return c.backupClient.StartCopyJob(ctx, params)
}

func (c *client) DescribeCopyJob(ctx context.Context, copyJobId string) (*backup.DescribeCopyJobOutput, error) {
	return c.backupClient.DescribeCopyJob(ctx, copyJobId)
}

// WebACL methods
func (c *client) ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error) {
	out, err := c.wafv2Client.ListWebACLs(ctx, &wafv2.ListWebACLsInput{
		Scope: scope,
	})
	if err != nil {
		return nil, err
	}
	return out.WebACLs, nil
}

func (c *client) GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error) {
	out, err := c.wafv2Client.GetWebACL(ctx, &wafv2.GetWebACLInput{
		Name:  &name,
		Id:    &id,
		Scope: scope,
	})
	if err != nil {
		return nil, "", err
	}
	lockToken := ""
	if out.LockToken != nil {
		lockToken = *out.LockToken
	}
	return out.WebACL, lockToken, nil
}

func (c *client) DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error {
	_, err := c.wafv2Client.DeleteWebACL(ctx, &wafv2.DeleteWebACLInput{
		Name:      &name,
		Id:        &id,
		Scope:     scope,
		LockToken: &lockToken,
	})
	return err
}

func (c *client) ListTagsForWebACL(ctx context.Context, resourceArn string) ([]wafv2types.Tag, error) {
	out, err := c.wafv2Client.ListTagsForResource(ctx, &wafv2.ListTagsForResourceInput{
		ResourceARN: &resourceArn,
	})
	if err != nil {
		return nil, err
	}
	if out.TagInfoForResource == nil {
		return nil, nil
	}
	return out.TagInfoForResource.TagList, nil
}

// Certificate methods
func (c *client) ListCertificates(ctx context.Context) ([]acmtypes.CertificateSummary, error) {
	out, err := c.acmClient.ListCertificates(ctx, &acm.ListCertificatesInput{})
	if err != nil {
		return nil, err
	}
	return out.CertificateSummaryList, nil
}

func (c *client) ListCertificateTags(ctx context.Context, arn string) ([]acmtypes.Tag, error) {
	out, err := c.acmClient.ListTagsForCertificate(ctx, &acm.ListTagsForCertificateInput{
		CertificateArn: &arn,
	})
	if err != nil {
		return nil, err
	}
	return out.Tags, nil
}

func (c *client) DeleteCertificate(ctx context.Context, arn string) error {
	_, err := c.acmClient.DeleteCertificate(ctx, &acm.DeleteCertificateInput{
		CertificateArn: &arn,
	})
	return err
}

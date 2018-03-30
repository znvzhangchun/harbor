package replication

import (
	"fmt"
	"net/http"

	common_http "github.com/vmware/harbor/src/common/http"
	"github.com/vmware/harbor/src/common/models"
	reg "github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/jobservice/env"
	"github.com/vmware/harbor/src/jobservice/logger"
)

// Replicator call UI's API to start a repliation according to the policy ID
// passed in parameters
type Replicator struct {
	ctx      env.JobContext
	url      string // the URL of UI service
	insecure bool
	policyID int64
	client   *common_http.Client
	logger   logger.Interface
}

// ShouldRetry ...
func (r *Replicator) ShouldRetry() bool {
	return false
}

// MaxFails ...
func (r *Replicator) MaxFails() uint {
	return 0
}

// Validate ....
func (r *Replicator) Validate(params map[string]interface{}) error {
	return nil
}

// Run ...
func (r *Replicator) Run(ctx env.JobContext, params map[string]interface{}) error {
	if err := r.init(ctx, params); err != nil {
		return err
	}
	return r.replicate()
}

func (r *Replicator) init(ctx env.JobContext, params map[string]interface{}) error {
	r.logger = ctx.GetLogger()
	r.ctx = ctx
	if canceled(r.ctx) {
		r.logger.Warning(errCanceled.Error())
		return errCanceled
	}

	r.policyID = (int64)(params["policy_id"].(float64))
	r.url = params["url"].(string)
	r.insecure = params["insecure"].(bool)
	cred := auth.NewCookieCredential(&http.Cookie{
		Name:  models.UISecretCookie,
		Value: secret(),
	})

	r.client = common_http.NewClient(&http.Client{
		Transport: reg.GetHTTPTransport(r.insecure),
	}, cred)

	r.logger.Infof("initialization completed: policy ID: %d, URL: %s, insecure: %v",
		r.policyID, r.url, r.insecure)

	return nil
}

func (r *Replicator) replicate() error {
	if err := r.client.Post(fmt.Sprintf("%s/api/replications", r.url), struct {
		PolicyID int64 `json:"policy_id"`
	}{
		PolicyID: r.policyID,
	}); err != nil {
		r.logger.Errorf("failed to send the replication request to %s: %v", r.url, err)
		return err
	}
	r.logger.Infof("the replication request has been sent to %s successfully", r.url)
	return nil

}

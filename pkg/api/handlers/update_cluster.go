package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/sapcc/kubernikus/pkg/api"
	"github.com/sapcc/kubernikus/pkg/api/models"
	"github.com/sapcc/kubernikus/pkg/api/rest/operations"
	"github.com/sapcc/kubernikus/pkg/apis/kubernikus/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func NewUpdateCluster(rt *api.Runtime) operations.UpdateClusterHandler {
	return &updateCluster{rt: rt}
}

type updateCluster struct {
	rt *api.Runtime
}

func (d *updateCluster) Handle(params operations.UpdateClusterParams, principal *models.Principal) middleware.Responder {

	_, err := editCluster(d.rt.Clients.Kubernikus.Kubernikus().Klusters(d.rt.Namespace), principal, params.Name, func(kluster *v1.Kluster) {
		//TODO: currently no field to update
	})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return NewErrorResponse(&operations.UpdateClusterDefault{}, 404, "Not found")
		}
		return NewErrorResponse(&operations.UpdateClusterDefault{}, 500, err.Error())
	}
	return operations.NewUpdateClusterOK()
}

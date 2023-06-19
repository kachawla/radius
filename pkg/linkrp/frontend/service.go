/*
Copyright 2023 The Radius Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package frontend

import (
	"context"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/project-radius/radius/pkg/armrpc/frontend/server"
	"github.com/project-radius/radius/pkg/armrpc/hostoptions"
	"github.com/project-radius/radius/pkg/linkrp/frontend/handler"

	ctrl "github.com/project-radius/radius/pkg/armrpc/frontend/controller"
)

type Service struct {
	server.Service
}

func NewService(options hostoptions.HostOptions) *Service {
	return &Service{
		server.Service{
			Options:      options,
			ProviderName: handler.ProviderNamespaceName,
		},
	}
}

func (s *Service) Name() string {
	return handler.ProviderNamespaceName
}

func (s *Service) Run(ctx context.Context) error {
	if err := s.Init(ctx); err != nil {
		return err
	}

	opts := ctrl.Options{
		Address:       fmt.Sprintf("%s:%d", s.Options.Config.Server.Host, s.Options.Config.Server.Port),
		PathBase:      s.Options.Config.Server.PathBase,
		DataProvider:  s.StorageProvider,
		KubeClient:    s.KubeClient,
		StatusManager: s.OperationStatusManager,
	}

	err := s.Start(ctx, server.Options{
		Address:           opts.Address,
		ProviderNamespace: s.ProviderName,
		Location:          s.Options.Config.Env.RoleLocation,
		PathBase:          s.Options.Config.Server.PathBase,
		// set the arm cert manager for managing client certificate
		ArmCertMgr:    s.ARMCertManager,
		EnableArmAuth: s.Options.Config.Server.EnableArmAuth, // when enabled the client cert validation will be done
		Configure: func(router *mux.Router) error {
			err := handler.AddRoutes(ctx, router, !hostoptions.IsSelfHosted(), opts)
			if err != nil {
				return err
			}
			return nil
		},
	})
	return err
}

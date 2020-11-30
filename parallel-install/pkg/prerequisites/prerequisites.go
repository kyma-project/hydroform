package prerequisites

import (
	"context"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
)

const logPrefix = "[prerequisites/prerequisites.go]"

func InstallPrerequisites(ctx context.Context, prerequisites []components.Component) <-chan error {

	statusChan := make(chan error)

	go func() {
		defer close(statusChan)

		for _, prerequisite := range prerequisites {
			//TODO: Is there a better way to find out if Context is canceled?
			if ctx.Err() != nil {
				//Context is canceled or timed-out. Skip processing
				config.Log("%s Finishing work: %v", logPrefix, ctx.Err())
				return
			}

			config.Log("%s Installing component %s ", logPrefix, prerequisite.Name)
			err := prerequisite.InstallComponent(ctx)
			if err != nil {
				statusChan <- err
				return
			}
			statusChan <- nil
		}
	}()

	return statusChan
}

func UninstallPrerequisites(ctx context.Context, prerequisites []components.Component) <-chan error {

	statusChan := make(chan error)

	go func() {
		defer close(statusChan)

		for i := len(prerequisites) - 1; i >= 0; i-- {
			prereq := prerequisites[i]
			//TODO: Is there a better way to find out if Context is canceled?
			if ctx.Err() != nil {
				//Context is canceled or timed-out. Skip processing
				config.Log("%s Finishing work: %v", logPrefix, ctx.Err())

				return
			}
			config.Log("%s Uninstalling component %s ", logPrefix, prereq.Name)

			err := prereq.UninstallComponent(ctx)
			if err != nil {
				statusChan <- err
				return
			}
			statusChan <- nil
		}
	}()

	return statusChan
}

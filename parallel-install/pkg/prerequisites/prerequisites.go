package prerequisites

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"log"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
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
			err := prerequisite.InstallComponent()
			if err != nil {
				statusChan <- err
				return
			}
			statusChan <- nil
		}
	}()

	return statusChan
}

func UninstallPrerequisites(ctx context.Context, prerequisites []components.Component) error {

	for i := len(prerequisites) - 1; i >= 0; i-- {
		prereq := prerequisites[i]
		//TODO: Is there a better way to find out if Context is canceled?
		if ctx.Err() != nil {
			//Context is canceled or timed-out. Skip processing
			return fmt.Errorf("Error uninstalling prerequisite %s: %v", prereq.Name, ctx.Err())
		}
		err := prereq.UninstallComponent(ctx)
		if err != nil {
			log.Printf("%s Error uninstalling prerequisite %s: %v (The uninstallation continues anyway)", logPrefix, prereq.Name, err)
		}
	}

	return nil
}

package stacker

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/containers/image/docker"
	"github.com/containers/image/types"
	"github.com/openSUSE/umoci"

	"github.com/anuvu/stacker/lib"
)

func getImageTags(baseUrl, name string) ([]string, error) {

	// Need to determine if URL is docker/oci or something else
	is, err := NewImageSource(baseUrl)
	if err != nil {
		return nil, err
	}

	switch is.Type {
	case DockerType:
		// Determine URL of image
		imageUrl := fmt.Sprintf("%s/%s", strings.TrimRight(baseUrl, "/"), name)

		// ToDo verify certificate
		systemCtx := &types.SystemContext{
			DockerInsecureSkipTLSVerify: types.OptionalBoolTrue,
		}

		imgRef, err := docker.ParseReference(imageUrl)
		if err != nil {
			return nil, err
		}

		return docker.GetRepositoryTags(context.Background(), systemCtx, imgRef)

	case OCIType:
		engine, err := umoci.OpenLayout(is.Url)
		if err != nil {
			return nil, err
		}
		defer engine.Close()

		refNames, err := engine.ListReferences(context.Background())
		if err != nil {
			return nil, err
		}

		var tagNames []string

		for _, refName := range refNames {
			strings.SplitN(refName, "_", 1)

		}


		imageUrl = fmt.Sprintf("%s:%s_%s", baseUrl, name, tag)
	default:
		return fmt.Errorf("can't search for tags from destination type: %s", is.Type)
	}



	return nil
}
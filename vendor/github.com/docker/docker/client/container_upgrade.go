package client

import (
	"strings"

	"github.com/docker/docker/api/types/container"
	"golang.org/x/net/context"
)

// ContainerUpgrade upgrade image of a container
func (cli *Client) ContainerUpgrade(ctx context.Context, containerID string, config container.Config) error {
	resp, err := cli.post(ctx, "/containers/"+containerID+"/upgrade", nil, config, nil)
	if err != nil {
		if resp.statusCode == 404 && strings.Contains(err.Error(), "No such image") {
			return imageNotFoundError{config.Image}
		}
		return err
	}

	ensureReaderClosed(resp)
	return err
}

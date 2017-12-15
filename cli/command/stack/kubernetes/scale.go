package kubernetes

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/docker/cli/kubernetes/compose/v1beta2"
)

func RunScale(dockerCli *KubeCli, opts []string) error {
	rest, err := dockerCli.restClientv1beta2()
	if err != nil {
		return err
	}
	var scaler v1beta2.Scale
	scaler.Name = opts[0]
	scaler.Spec = make(map[string]int)
	for i:=1; i<len(opts); i++ {
		nv := strings.Split(opts[i], "=")
		if len(nv) != 2 {
			return fmt.Errorf("argument '%s' is not of the form 'name=count'", opts[i])
		}
		i, err := strconv.Atoi(nv[1])
		if err != nil || i < 0 {
			return fmt.Errorf("'%s' is not a positive integer: %v", nv[1], err)
		}
		scaler.Spec[nv[0]] = i
	}
	err = rest.Put().Namespace(dockerCli.kubeNamespace).Name(opts[0]).Resource("stacks").SubResource("scale").Body(&scaler).Do().Error()
	return err
}

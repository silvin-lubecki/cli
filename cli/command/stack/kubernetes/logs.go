package kubernetes

import (
	"bufio"
	"hash/crc32"
	"io"
	"fmt"
	"strings"
)

func RunLogs(dockerCli *KubeCli, opts []string, colorized bool, follow bool, tail int) error {
	rest, err := dockerCli.restClientv1beta2()
	if err != nil {
		return err
	}
	req := rest.Get().Namespace(dockerCli.kubeNamespace).Name(opts[0]).Resource("stacks").SubResource("log")
	if follow {
		req = req.Param("follow", "true")
	}
	if tail >= 0 {
		req = req.Param("tail", fmt.Sprintf("%v", tail))
	}
	rc, err := req.Stream()
	if err != nil {
		return err
	}
	defer rc.Close()
	if !colorized {
		io.Copy(dockerCli.Out(), rc)
	} else {
		reader := bufio.NewReader(rc)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			sep := strings.Index(line, " ")
			if sep == -1 {
				sep = len(line)
			}
			hash := crc32.ChecksumIEEE([]byte(line[0:sep]))
			dockerCli.Out().Write([]byte(fmt.Sprintf("\x1b[%dm%s\x1b[0m%s", 29 + (hash%9), line[0:sep], line[sep:])))
		}
	}
	return nil
}

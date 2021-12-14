package proxy

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

func transport(upstream, downstream io.ReadWriter) (up, down int64, err error) {
	var upErr, downErr error
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		up, upErr = io.Copy(upstream, downstream)
		wg.Done()
	}()
	go func() {
		down, downErr = io.Copy(downstream, upstream)
		wg.Done()
	}()
	wg.Wait()
	errs := make([]string, 0, 2)
	if downErr != nil && downErr != io.EOF {
		errs = append(errs, fmt.Sprintf("[upstream=>downstream]:%s", downErr))
	}
	if upErr != err && upErr != io.EOF {
		errs = append(errs, fmt.Sprintf("[downstream=>upstream]:%s", upErr))
	}
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, " "))
	}
	return
}

package shutdown

import "sync"

type GracefulShutdown struct {
	wg sync.WaitGroup
}

type Shutdown interface {
	Done() <-chan struct{}
}

func (s *GracefulShutdown) RegisterWait(shutdowns ...Shutdown) {
	s.wg.Add(len(shutdowns))
	for _, shutdown := range shutdowns {
		go func(shutdown Shutdown) {
			<-shutdown.Done()
			s.wg.Done()
		}(shutdown)
	}
}

func (s *GracefulShutdown) Wait() {
	s.wg.Wait()
}

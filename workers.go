package main

const STOP_WORKERS = false

func allAliveWorkers() bool {

	var allAlive = true

	processes.Range(func(_, process any) bool {
		if !process.(Goroutine).Alive {
			allAlive = false
		}
		return true
	})

	return allAlive
}

func stopWorkers() {

	processes.Range(func(processId, process any) bool {
		pr := process.(Goroutine)
		pr.Alive = false

		processes.Store(processId, pr)
		return true
	})
}

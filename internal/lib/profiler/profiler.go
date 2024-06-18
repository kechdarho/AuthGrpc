package profiler

import (
	"os"
	"runtime/pprof"
)

type ProfilerConfig struct {
	CPUProfilePath string
	MemProfilePath string
}

type Profiler struct {
	config  ProfilerConfig
	cpuFile *os.File
	memFile *os.File
}

func NewProfiler(config ProfilerConfig) *Profiler {
	return &Profiler{config: config}
}

func (p *Profiler) StartCPUProfile() error {
	var err error
	p.cpuFile, err = os.Create(p.config.CPUProfilePath)
	if err != nil {
		return err
	}
	err = pprof.StartCPUProfile(p.cpuFile)
	if err != nil {
		return err
	}
	return nil
}

func (p *Profiler) StopCPUProfile() {
	if p.cpuFile != nil {
		pprof.StopCPUProfile()
		p.cpuFile.Close()
		if err := p.cpuFile.Close(); err != nil {
			panic(err)
		}
	}
}

func (p *Profiler) WriteHeapProfile() error {
	var err error
	p.memFile, err = os.Create(p.config.MemProfilePath)
	if err != nil {
		return err
	}
	defer func(memFile *os.File) {
		err := memFile.Close()
		if err != nil {

		}
	}(p.memFile)
	return pprof.WriteHeapProfile(p.memFile)
}

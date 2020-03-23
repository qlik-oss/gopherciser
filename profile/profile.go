package profile

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	ProfileType int
	profile     struct {
		typ        ProfileType
		closeFuncs []func() error
	}
)

const (
	ProfileTypeUnknown ProfileType = iota // Default value of int
	ProfileTypeCPU
	ProfileTypeBlock
	ProfileTypeGoRoutine
	ProfileTypeTreadCreate
	ProfileTypeHeap
	ProfileTypeMutex
	ProfileTypeTrace
	ProfileTypeMem
)

var (
	// ProfileTypeEnum enum of profiling options
	ProfileTypeEnum, _ = enummap.NewEnumMap(map[string]int{
		"cpu":          int(ProfileTypeCPU),
		"block":        int(ProfileTypeBlock),
		"goroutine":    int(ProfileTypeGoRoutine),
		"threadcreate": int(ProfileTypeTreadCreate),
		"heap":         int(ProfileTypeHeap),
		"mutex":        int(ProfileTypeMutex),
		"trace":        int(ProfileTypeTrace),
		"mem":          int(ProfileTypeMem),
	})

	profiler *profile
)

// Close profiler, registered close functions will be executed LIFO
func Close() error {
	if profiler == nil {
		return nil
	}

	var multiErr *multierror.Error
	for i := len(profiler.closeFuncs) - 1; i >= 0; i-- {
		if err := profiler.closeFuncs[i](); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return errors.WithStack(helpers.FlattenMultiError(multiErr))
}

func validateProfiler() error {
	if profiler != nil {
		return errors.Errorf("Profiler already set to %s, only one profer can be used at a time",
			ProfileTypeEnum.StringDefault(int(profiler.typ), "unknown"))
	}
	return nil
}

// StartCPUProfiler and save to file
func StartCPUProfiler(file string) error {
	if err := validateProfiler(); err != nil {
		return errors.Wrap(err, "failed to start cpu profiler")
	}

	cpuProfile, err := os.Create(file)
	if err != nil {
		return errors.Wrapf(err, "failed to create cpu profiling file<%s>", file)
	}

	profiler = &profile{typ: ProfileTypeCPU}
	profiler.closeFuncs = append(profiler.closeFuncs, cpuProfile.Close)

	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		return errors.Wrap(err, "failed to start cpu profiler")
	}
	profiler.closeFuncs = append(profiler.closeFuncs, func() error { pprof.StopCPUProfile(); return nil })

	return nil
}

// StartBlockProfiler and save to file, debug flag adds stacks to profile
func StartBlockProfiler(file string, debug int) error {
	if err := validateProfiler(); err != nil {
		return errors.Wrap(err, "failed to start block profiler")
	}

	blockProfile, err := os.Create(file)
	if err != nil {
		return errors.Wrap(err, "Failed to create block profiling file")
	}

	profiler = &profile{typ: ProfileTypeBlock}
	runtime.SetBlockProfileRate(1)

	profiler.closeFuncs = append(profiler.closeFuncs, blockProfile.Close)
	profiler.closeFuncs = append(profiler.closeFuncs, func() error { return pprof.Lookup("block").WriteTo(blockProfile, debug) })

	return nil
}

// StartGoroutineProfiler and save to a fileprefix_%i.out file each t
func StartGoroutineProfiler(fileprefix string, t time.Duration) error {
	if err := validateProfiler(); err != nil {
		return errors.Wrap(err, "failed to start goroutine profiler")
	}

	profiler = &profile{typ: ProfileTypeGoRoutine}

	ticker := time.NewTicker(t)
	i := 0
	go func() {
		for range ticker.C {
			i++
			func() {
				goroutineProfile, err := os.Create(fmt.Sprintf("%s_%d.out", fileprefix, i))
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", errors.Wrap(err, "Failed to create goroutine profiling file"))
				}

				defer func() {
					if err := goroutineProfile.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "failed to close file<%s>\n", goroutineProfile.Name())
					}
				}()

				defer func() {
					if err := pprof.Lookup("goroutine").WriteTo(goroutineProfile, 0); err != nil {
						fmt.Fprintf(os.Stderr, "error writing goroutine profile to %s: %v\n", goroutineProfile.Name(), err)
					}
				}()
			}()
		}
	}()
	profiler.closeFuncs = append(profiler.closeFuncs, func() error { ticker.Stop(); return nil })
	return nil
}

// StartThreadCreateProfiler and save to file
func StartThreadCreateProfiler(file string) error {
	if err := validateProfiler(); err != nil {
		return errors.Wrap(err, "failed to start threadcreate profiler")
	}

	profiler = &profile{typ: ProfileTypeTreadCreate}

	tcProfile, err := os.Create(file)
	if err != nil {
		return errors.Wrapf(err, "failed to create threadcreate profiling file<%s>", file)
	}

	profiler.closeFuncs = append(profiler.closeFuncs, tcProfile.Close)
	profiler.closeFuncs = append(profiler.closeFuncs, func() error { return pprof.Lookup("threadcreate").WriteTo(tcProfile, 0) })

	return nil
}

// StartHeapProfiler and save to file
func StartHeapProfiler(file string) error {
	if err := validateProfiler(); err != nil {
		return errors.Wrap(err, "failed to start heap profiler")
	}

	profiler = &profile{typ: ProfileTypeHeap}

	heapProfile, err := os.Create(file)
	if err != nil {
		return errors.Wrapf(err, "failed to create heap profiling file<%s>", file)
	}

	profiler.closeFuncs = append(profiler.closeFuncs, heapProfile.Close)
	profiler.closeFuncs = append(profiler.closeFuncs, func() error { return pprof.Lookup("heap").WriteTo(heapProfile, 0) })
	return nil
}

// StartMutexProfiler and save to file, debug flag adds stacks to profile
func StartMutexProfiler(file string, debug int) error {
	if err := validateProfiler(); err != nil {
		return errors.Wrap(err, "failed to start mutex profiler")
	}

	profiler = &profile{typ: ProfileTypeMutex}

	runtime.SetMutexProfileFraction(1)
	mutexProfile, err := os.Create(file)
	if err != nil {
		return errors.Wrapf(err, "failed to create mutex profiling file<%s>", file)
	}

	profiler.closeFuncs = append(profiler.closeFuncs, mutexProfile.Close)
	profiler.closeFuncs = append(profiler.closeFuncs, func() error { return pprof.Lookup("mutex").WriteTo(mutexProfile, debug) })

	return nil
}

// StartTraceProfiler and save to file
func StartTraceProfiler(file string) error {
	if err := validateProfiler(); err != nil {
		return errors.Wrap(err, "failed to start trace profiler")
	}

	profiler = &profile{typ: ProfileTypeTrace}

	traceFile, err := os.Create(file)
	if err != nil {
		return errors.Wrapf(err, "failed to create trace file<%s>", file)
	}
	profiler.closeFuncs = append(profiler.closeFuncs, traceFile.Close)

	if err = trace.Start(traceFile); err != nil {
		return errors.Wrap(err, "failed to start trace")
	}
	profiler.closeFuncs = append(profiler.closeFuncs, func() error { trace.Stop(); return nil })

	return nil
}

// StartMemProfiler and save to file
func StartMemProfiler(file string) error {
	if err := validateProfiler(); err != nil {
		return errors.Wrap(err, "failed to start mem profiler")
	}

	profiler = &profile{typ: ProfileTypeTrace}

	runtime.MemProfileRate = 1

	memProfile, err := os.Create(file)
	if err != nil {
		return errors.Wrapf(err, "failed to create mem profiling file<%s>", file)
	}

	profiler.closeFuncs = append(profiler.closeFuncs, memProfile.Close)
	profiler.closeFuncs = append(profiler.closeFuncs, func() error { return pprof.WriteHeapProfile(memProfile) })

	return nil
}

// StartProfiler start profiler of type (typ) with default parameters
func StartProfiler(typ ProfileType) error {
	switch typ {
	case ProfileTypeBlock:
		return errors.WithStack(StartBlockProfiler("block.out", 1))
	case ProfileTypeCPU:
		return errors.WithStack(StartCPUProfiler("cpu.out"))
	case ProfileTypeGoRoutine:
		return errors.WithStack(StartGoroutineProfiler("goroutine", time.Second))
	case ProfileTypeHeap:
		return errors.WithStack(StartHeapProfiler("heap.out"))
	case ProfileTypeMem:
		return errors.WithStack(StartMemProfiler("mem.out"))
	case ProfileTypeMutex:
		return errors.WithStack(StartMutexProfiler("mutex.out", 1))
	case ProfileTypeTrace:
		return errors.WithStack(StartTraceProfiler("trace.out"))
	case ProfileTypeTreadCreate:
		return errors.WithStack(StartThreadCreateProfiler("threadcreate.out"))
	default:
		return errors.Errorf("Unsupported profile type<%s>", ProfileTypeEnum.StringDefault(int(typ), strconv.Itoa(int(typ))))
	}
}

// ResolveParameter of string or int to profiletype
func ResolveParameter(param string) (ProfileType, error) {
	var i int
	var err error

	// Do we have an int?
	if i, err = strconv.Atoi(param); err != nil {
		// it's a string
		i, err = ProfileTypeEnum.Int(param)
		if err != nil {
			return ProfileTypeUnknown, errors.Wrapf(err, "failed to parse <%s> to profile type", param)
		}
	}
	// it's an int

	// make sure it's a valid type
	_, err = ProfileTypeEnum.String(i)
	if err != nil {
		return ProfileTypeUnknown, errors.Wrapf(err, "failed to parse <%s> to profile type", param)
	}

	return ProfileType(i), nil
}

// Help formats a string to be displayed as help
func Help() string {
	buf := helpers.NewBuffer()

	buf.WriteString("Set a profiler to be started. One of:\n")
	ProfileTypeEnum.ForEachSorted(func(k int, v string) {
		buf.WriteString("[")
		buf.WriteString(strconv.Itoa(k))
		buf.WriteString("]: ")
		buf.WriteString(v)
		buf.WriteString("\n")
	})
	return buf.String()
}

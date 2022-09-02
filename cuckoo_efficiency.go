package main

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/GuoxinL/cuckoo-efficiency/report"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/linvon/cuckoo-filter"
	"github.com/panjf2000/ants/v2"
)

var (
	keys              [][]byte
	supplementaryKeys [][]byte
)

func efficiency(size, height int, b, f, typ uint, goroutinesNumber, loopNumber int, reportPath string) {
	var (
		charters []components.Charter
	)
	realSize := GetApproximationMaxNumKeys(uint32(size), uint32(b))
	filter := cuckoo.NewFilter(b, f, realSize, typ)
	if len(keys) != size {
		keys = nil
		for i := 0; i < size; i++ {
			time.Sleep(time.Nanosecond)
			keys = append(keys, GenTimestampKey())
		}
	}
	if len(supplementaryKeys) != size {
		supplementaryKeys = nil
		for i := 0; i < size; i++ {
			time.Sleep(time.Nanosecond)
			supplementaryKeys = append(supplementaryKeys, GenTimestampKey())
		}
	}

	heights := make([]uint64, 0, size/height)
	addCosts := make([]interface{}, 0, size/height)
	containsCosts := make([]interface{}, 0, size/height)
	fpCount := make([]interface{}, 0, size/height)

	for i := 0; i < size/height; i++ {
		var fp uint32
		// 查重
		now := time.Now()
		for j := 0; j < height; j++ {
			key := keys[i*height+j]
			contains := filter.Contain(key)
			if contains {
				fp++
				continue
			}
		}
		containSince := time.Since(now)

		// 添加
		now = time.Now()
		for j := 0; j < height; j++ {
			key := keys[i*height+j]
			add := filter.Add(key)
			if !add {
				fmt.Println("is full")
				break
			}
		}
		addSince := time.Since(now)
		// 记录时长消耗
		addCosts = append(addCosts, addSince.Nanoseconds())
		containsCosts = append(containsCosts, containSince.Nanoseconds())
		heights = append(heights, uint64(i))
		fpCount = append(fpCount, fp)
		fmt.Println(fmt.Sprintf("[%v], stored:%v, FP: %v, contains: %v, add: %v",
			i, filter.Size(), fp, containSince, addSince))
	}

	encode, _ := filter.Encode()

	line := report.Report("cuckoo filter contains and add",
		fmt.Sprintf("Size:%v,Full Size:%v,Real Size:%v,TableType:%v,TagsPerBucket:%v,BitsPerItem:%v,Pool:%v,Space:%v",
			size, filter.Size(), realSize, typ, b, f, goroutinesNumber,
			strconv.FormatFloat(float64(len(encode))/1024/1024, 'f', 2, 64)+"mb"),
		heights,
		report.Series{Name: "Add Costs", Data: addCosts},
		report.Series{Name: "Contains Costs", Data: containsCosts},
	)
	charters = append(charters, line)

	line = report.Report("cuckoo filter false positive",
		fmt.Sprintf("Size:%v,Full Size:%v,Real Size:%v,TableType:%v,TagsPerBucket:%v,BitsPerItem:%v,Pool:%v,Space:%v",
			size, filter.Size(), realSize, typ, b, f, goroutinesNumber,
			strconv.FormatFloat(float64(len(encode))/1024/1024, 'f', 2, 64)+"mb"),
		heights,
		report.Series{Name: "FP Count", Data: fpCount},
	)
	charters = append(charters, line)

	for i := 0; i < loopNumber; i++ {
		cuckooSupplementaryLine, cuckooFPSupplementaryLine := parallelContainsAndReport(size, height, filter, b, f, typ,
			realSize, goroutinesNumber, i+1, loopNumber)
		charters = append(charters, cuckooSupplementaryLine, cuckooFPSupplementaryLine)
	}

	report.GeneratePages(
		fmt.Sprintf("cuckoo filter Size:%v,Full Size:%v,Real Size:%v,TableType:%v,TagsPerBucket:%v,BitsPerItem:%v,",
			size, filter.Size(), realSize, typ, b, f), reportPath, charters...)
}

func parallelContainsAndReport(size, height int, filter *cuckoo.Filter, b, f, typ, realSize uint, goroutinesNumber,
	number, loopNumber int) (*charts.Line, *charts.Line) {

	supplementaryContainsCosts := make([]interface{}, size/height)
	supplementaryFPCount := make([]interface{}, size/height)
	supplementaryHeights := make([]uint64, size/height)

	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(goroutinesNumber, func(i interface{}) {
		parallelContains(size, height, filter, supplementaryContainsCosts, supplementaryFPCount, supplementaryHeights)
		wg.Done()
	})
	defer p.Release()
	now := time.Now()
	// Submit tasks one by one.
	for i := 0; i < size/height; i++ {
		wg.Add(1)
		_ = p.Invoke(int32(i))
	}
	wg.Wait()
	totalCosts := time.Since(now)
	cuckooSupplementaryLine := report.Report(
		fmt.Sprintf("cuckoo filter parallel contains test %v/%v", loopNumber, number),
		fmt.Sprintf("Size:%v,Full Size:%v,Real Size:%v,TableType:%v,TagsPerBucket:%v,BitsPerItem:%v,Pool:%v,"+
			"TotalCosts:%v",
			size, filter.Size(), realSize, typ, b, f, goroutinesNumber, totalCosts,
		),
		supplementaryHeights,
		report.Series{Name: "Contains Costs", Data: supplementaryContainsCosts},
	)

	cuckooFPSupplementaryLine := report.Report(fmt.Sprintf("cuckoo filter contains false positive %v/%v",
		loopNumber, number),
		fmt.Sprintf("Size:%v,Full Size:%v,Real Size:%v,TableType:%v,TagsPerBucket:%v,BitsPerItem:%v,Pool:%v",
			size, filter.Size(), realSize, typ, b, f, goroutinesNumber,
		),
		supplementaryHeights,
		report.Series{Name: "FP Count", Data: supplementaryFPCount},
	)
	return cuckooSupplementaryLine, cuckooFPSupplementaryLine
}

func parallelContains(size, height int, filter *cuckoo.Filter, costs, fpCount []interface{}, heights []uint64) {
	for i := 0; i < size/height; i++ {
		var fp uint32
		// 查重
		now := time.Now()
		for j := 0; j < height; j++ {
			key := supplementaryKeys[i*height+j]
			contains := filter.Contain(key)
			if contains {
				fp++
				continue
			}
		}
		since := time.Since(now)
		costs[i] = since.Nanoseconds()
		fpCount[i] = fp
		heights[i] = uint64(i)
	}
}

func GetApproximationMaxNumKeys(maxNumKeys, b uint32) uint {
	loadFactor := GetLoadFactor(b)
	got := float64(maxNumKeys) * 1.25 / loadFactor
	for i := float64(1); true; i++ {
		pow := math.Pow(2, i)
		rl := pow * loadFactor
		if rl > got {
			return uint(rl)
		}
	}
	return uint(maxNumKeys)
}

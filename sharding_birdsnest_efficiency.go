/*
   Created by guoxin in 2022/9/5 11:54 AM
*/
package eff

import (
	"chainmaker.org/chainmaker/common/v2/birdsnest"
	"fmt"
	"github.com/GuoxinL/cuckoo-efficiency/report"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/linvon/cuckoo-filter"
	"github.com/panjf2000/ants/v2"
	"runtime"
	"sync"
	"testing"
	"time"
)

func shardingBNEfficiency(size, height int, b, f, typ uint, goroutinesNumber, loopNumber int, reportPath string, t *testing.T) {
	var (
		charters []components.Charter
	)
	filter, err := birdsnest.NewBirdsNest(&birdsnest.BirdsNestConfig{
		ChainId: "chain1",
		Length:  10,
		Rules:   &birdsnest.RulesConfig{AbsoluteExpireTime: 1000000},
		Cuckoo: &birdsnest.CuckooConfig{
			KeyType:       birdsnest.KeyType_KTTimestampKey,
			TagsPerBucket: uint32(b),
			BitsPerItem:   uint32(f),
			MaxNumKeys:    100_000,
			TableType:     cuckoo.TableTypeSingle,
		},
		Snapshot: &birdsnest.SnapshotSerializerConfig{
			Type:  birdsnest.SerializeIntervalType_Timed,
			Timed: &birdsnest.TimedSerializeIntervalConfig{Interval: 1},
			Path:  "./data",
		},
	}, make(chan struct{}), birdsnest.LruStrategy, birdsnest.TestLogger{T: t})
	if err != nil {
		fmt.Println(err)
		return
	}
	//filter := cuckoo.NewFilter(b, f, realSize, typ)
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
			contains, _, err := filter.Contains(birdsnest.TimestampKey(key))
			if err != nil {
				fmt.Println("contains error: ", err)
				continue
			}
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
			err := filter.Add(birdsnest.TimestampKey(key))
			if err != nil {
				fmt.Println("contains error: ", err)
				return
			}
		}
		addSince := time.Since(now)
		// 记录时长消耗
		addCosts = append(addCosts, addSince.Nanoseconds())
		containsCosts = append(containsCosts, containSince.Nanoseconds())
		heights = append(heights, uint64(i))
		fpCount = append(fpCount, fp)
		fmt.Println(fmt.Sprintf("[%v], FP: %v, contains: %v, add: %v",
			i, fp, containSince, addSince))
	}

	line := report.Report("Bird's nest contains and add",
		fmt.Sprintf("Size:%v,TableType:%v,TagsPerBucket:%v,BitsPerItem:%v,Pool:%v,",
			size, typ, b, f, goroutinesNumber),
		heights,
		report.Series{Name: "Add Costs", Data: addCosts},
		report.Series{Name: "Contains Costs", Data: containsCosts},
	)
	charters = append(charters, line)

	line = report.Report("Bird's nest false positive",
		fmt.Sprintf("Size:%v,TableType:%v,TagsPerBucket:%v,BitsPerItem:%v,Pool:%v",
			size, typ, b, f, goroutinesNumber),
		heights,
		report.Series{Name: "FP Count", Data: fpCount},
	)
	charters = append(charters, line)

	loopNumbers := make([]uint64, 0, loopNumber)
	totalCostses := make([]interface{}, 0, loopNumber)
	fpCount = make([]interface{}, 0, loopNumber)

	for i := 0; i < loopNumber; i++ {
		cuckooSupplementaryLine, cuckooFPSupplementaryLine, totalCosts, totalFPCount := parallelShardingBNContainsAndReport(size, height, filter, b, f, typ, goroutinesNumber, i+1, loopNumber)
		charters = append(charters, cuckooSupplementaryLine, cuckooFPSupplementaryLine)
		loopNumbers = append(loopNumbers, uint64(i))
		totalCostses = append(totalCostses, totalCosts.Nanoseconds())
		fpCount = append(fpCount, totalFPCount)
	}

	report.GeneratePages(
		fmt.Sprintf("Bird's nest Size:%v,TableType:%v,TagsPerBucket:%v,BitsPerItem:%v,",
			size, typ, b, f), reportPath, charters...)
	charters = nil
	line = report.Report("Bird's nest total costs", "", loopNumbers,
		report.Series{Name: "total Costs", Data: totalCostses},
	)
	charters = append(charters, line)
	line = report.Report("Bird's nest total false positive", "", loopNumbers,
		report.Series{Name: "total FP", Data: fpCount},
	)
	charters = append(charters, line)
	report.GeneratePages("Bird's nest total costs and FP", reportPath, charters...)
	runtime.GC()
}

//
func parallelShardingBNContainsAndReport(size, height int, bn birdsnest.BirdsNest, b, f, typ uint, goroutinesNumber, number, loopNumber int) (*charts.Line, *charts.Line, time.Duration, uint32) {

	supplementaryContainsCosts := make([]interface{}, size/height)
	supplementaryFPCount := make([]interface{}, size/height)
	supplementaryHeights := make([]uint64, size/height)

	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(goroutinesNumber, func(i interface{}) {
		parallelShardingBNContains(size, height, bn, supplementaryContainsCosts, supplementaryFPCount, supplementaryHeights)
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
		fmt.Sprintf("Bird's nest parallel contains test %v/%v", loopNumber, number),
		fmt.Sprintf("Size:%v,TableType:%v,TagsPerBucket:%v,BitsPerItem:%v,Pool:%v,"+
			"TotalCosts:%v",
			size, typ, b, f, goroutinesNumber, totalCosts,
		),
		supplementaryHeights,
		report.Series{Name: "Contains Costs", Data: supplementaryContainsCosts},
	)

	cuckooFPSupplementaryLine := report.Report(fmt.Sprintf("Bird's nest contains false positive %v/%v",
		loopNumber, number),
		fmt.Sprintf("Size:%v,TableType:%v,TagsPerBucket:%v,BitsPerItem:%v,Pool:%v",
			size, typ, b, f, goroutinesNumber),
		supplementaryHeights,
		report.Series{Name: "FP Count", Data: supplementaryFPCount},
	)
	var totalFPCount uint32
	for _, i := range supplementaryFPCount {
		totalFPCount += i.(uint32)
	}

	return cuckooSupplementaryLine, cuckooFPSupplementaryLine, totalCosts, totalFPCount
}

func parallelShardingBNContains(size, height int, bn birdsnest.BirdsNest, costs, fpCount []interface{}, heights []uint64) {
	for i := 0; i < size/height; i++ {
		var fp uint32
		// 查重
		now := time.Now()
		for j := 0; j < height; j++ {
			key := supplementaryKeys[i*height+j]
			contains, _, err := bn.Contains(birdsnest.TimestampKey(key))
			if err != nil {
				fmt.Println("contains error: ", err)
				continue
			}
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

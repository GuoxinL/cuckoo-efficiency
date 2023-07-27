/*
   Created by guoxin in 2022/4/10 12:47 PM
*/
package eff

import (
	"flag"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/linvon/cuckoo-filter"
)

var fingerprint = flag.Uint("fingerprint", 16, "指纹长度，用于\"压力性能分析\"")
var tableType = flag.Uint("table-type", cuckoo.TableTypeSingle, "过滤器类型 0 Single, 1 Packed")
var size = flag.Int("size", SIZE, "测试布谷鸟过滤器的总数据量")
var height = flag.Int("height", HEIGHT, "高度：将总数据量/高度=每个批次处理的数据")
var goroutines = flag.Int("goroutines", ContainsGoroutinesNumber, "contains操作，并发的协程池数量")
var loop = flag.Int("loop", 1, "第二轮压测循环次数")
var reportPath = flag.String("report-path", "./test_report_"+time.Now().Format("2006-01-02_15:04:05"),
	"第二轮压测循环次数")

func printParameters() {
	fmt.Println("init parameters")
	fmt.Println(
		fmt.Sprintf("--fingerprint=%v", *fingerprint),
		fmt.Sprintf("--table-type=%v", *tableType),
		fmt.Sprintf("--size=%v", *size),
		fmt.Sprintf("--height=%v", *height),
		fmt.Sprintf("--goroutines=%v", *goroutines),
		fmt.Sprintf("--loop=%v", *loop),
		fmt.Sprintf("--report-path=%v", *reportPath),
	)
	fmt.Println()
}

const (
	SIZE                     = 1_000_000
	HEIGHT                   = 10_000
	ContainsGoroutinesNumber = 100
)

// 穷举配置性能分析
func Test_exhaustion(t *testing.T) {
	printParameters()
	var tests []cas
	for i := float64(1); i <= 4; i++ {
		b := uint(math.Pow(2, i))
		for f := uint(9); f <= 32; f++ {
			case0 := cas{
				args: args{
					size:             *size,
					height:           *height,
					b:                b,
					f:                f,
					typ:              *tableType,
					goroutinesNumber: *goroutines,
					loopNumber:       *loop,
				},
			}
			case0.name = fmt.Sprintf("性能测试 size:%v,height:%v,b:%v,f:%v,type:%v,goroutinesNumber:%v,loopNumber:%v",
				case0.args.size, case0.args.height, case0.args.b, case0.args.f, case0.args.typ, case0.args.goroutinesNumber, case0.args.loopNumber)
			tests = append(tests, case0)
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			efficiency(tt.args.size, tt.args.height, tt.args.b, tt.args.f, tt.args.typ, tt.args.goroutinesNumber,
				tt.args.loopNumber, *reportPath)
		})
	}
}

func Test_pressure(t *testing.T) {
	printParameters()
	var tests []cas
	for i := float64(1); i <= 4; i++ {
		b := uint(math.Pow(2, i))
		case0 := cas{
			args: args{
				size:             *size,
				height:           *height,
				b:                b,
				f:                *fingerprint,
				typ:              *tableType,
				goroutinesNumber: *goroutines,
				loopNumber:       *loop,
			},
		}
		case0.name = fmt.Sprintf("性能测试 size:%v,height:%v,b:%v,f:%v,type:%v,goroutinesNumber:%v,loopNumber:%v",
			case0.args.size, case0.args.height, case0.args.b, case0.args.f, case0.args.typ, case0.args.goroutinesNumber, case0.args.loopNumber)
		tests = append(tests, case0)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			efficiency(tt.args.size, tt.args.height, tt.args.b, tt.args.f, tt.args.typ, tt.args.goroutinesNumber,
				tt.args.loopNumber, *reportPath)
		})
	}
}

func Test_b2_f11(t *testing.T) {
	printParameters()
	var tests []cas
	tests = append(tests, cas{
		name: fmt.Sprintf("性能测试 size:%v,height:%v,b:%v,f:%v,type:%v,goroutinesNumber:%v,loopNumber:%v",
			*size, *height, 2, *fingerprint, *tableType, *goroutines, *loop),
		args: args{
			size:             *size,
			height:           *height,
			b:                2,
			f:                32,
			typ:              *tableType,
			goroutinesNumber: *goroutines,
			loopNumber:       *loop,
		},
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			efficiency(tt.args.size, tt.args.height, tt.args.b, tt.args.f, tt.args.typ, tt.args.goroutinesNumber,
				tt.args.loopNumber, *reportPath)
		})
	}
}

func Test_buckedLen(t *testing.T) {
	var (
		f    = 11
		b    = 2
		size = 1_000_000
	)

	i := f*b*size + 7
	i = i >> 3
	fmt.Println(i)
}

type args struct {
	size             int
	height           int
	b                uint
	f                uint
	typ              uint
	goroutinesNumber int
	loopNumber       int
}

type cas struct {
	name string
	args args
}

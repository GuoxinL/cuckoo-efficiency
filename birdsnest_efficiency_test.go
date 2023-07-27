/*
   Created by guoxin in 2022/9/5 7:19 PM
*/
package eff

import (
	"fmt"
	"testing"
)

func Test_bnEfficiency(t *testing.T) {
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
			bnEfficiency(tt.args.size, tt.args.height, tt.args.b, tt.args.f, tt.args.typ, tt.args.goroutinesNumber,
				tt.args.loopNumber, *reportPath, t)
		})
	}
}

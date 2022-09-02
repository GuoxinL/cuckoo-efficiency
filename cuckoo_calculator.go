/*
   Created by guoxin in 2022/4/1 11:26 AM
*/
package main

import (
	"fmt"
	"math"
	"strconv"
)

const (
	one float64 = 1
	tow float64 = 2

	DefaultLoadFactor = 0.98
)

func main() {
	fmt.Println("init...")
	var (
		// b 桶大小
		b float64 = 2
		// f 指纹大小
		f float64 = 32
		// α 负载因子 nolint
		α float64 = 0
		// C 每个项均摊成本
		C float64 = 0
		// 指定过滤器大小
		size float64 = 2_000_000
	)
	α = GetLoadFactor(uint32(b))
	fmt.Println("init success")

	fmt.Println("b: ", b, " f: ", f, " α: ", α)
	fmt.Println("计算假阳性概率 2b/2^f< r")
	// 为了保证假阳性率 r，需要保证 2b/2^f< r
	r := tow * b / math.Pow(2, f)
	printFloat(r)

	fmt.Println("指纹长度检查")
	fmt.Println("计算指纹应保证 \"f ≥ log_2(2b/r)=log_2(1/r) + log_2(2b)\"")
	f1 := math.Log2(tow * b / r)
	f2 := math.Log2(1/r) + math.Log2(tow*b)
	Ok("fingerprint %v >= %v && %v == %v", f >= f1 && f1 == f2, f, f1, f1, f2)

	fmt.Println("计算项均摊成本")
	fmt.Println("计算每个项的均摊成本应保证 \"C ≤ [log_2(1/r) + log_2(2b)]/α\"")
	C = (math.Log2(one/r) + math.Log2(tow*b)) / α
	fmt.Println(fmt.Sprintf("C 每个项均摊成本: \"C ≤ %v\"", C))

	fmt.Println()
	fmt.Println("半排序桶占用成本检查")
	fmt.Println("使用半排序时，应保证 \"ceil(b*(f-1)/8)<ceil(b*f/8)\" 否则是否使用半排序占用的空间是一样大的")
	up := math.Ceil(b * (f - 1) / 8)
	low := math.Ceil(b * f / 8)
	Ok("半排序桶 %v < %v", up < low, up, low)

	fmt.Println("过滤器大小选择")
	fmt.Println("过滤器的桶总大小一定是 2 的指数倍，因此在设定过滤器大小时，尽量满足 size/α ~=(<) 2^n，size 即为")
	fmt.Println("想要一个过滤器存储的数据量，必要时应选择小一点的过滤器，使用多个过滤器达到目标效果")
	got := size * 1.25 / GetLoadFactor(uint32(b))
	var approximation float64
	for i := float64(1); true; i++ {
		pow := math.Pow(2, i)
		rl := pow * GetLoadFactor(uint32(b)) * 0.8
		fmt.Println("幂的次数I:", i, " 2的I次幂: ", strconv.FormatFloat(pow, 'f', -1, 64), "got:",
			strconv.FormatFloat(got, 'f', -1, 64), "近似值: ", strconv.FormatFloat(rl, 'f',
				-1, 64))
		if rl > got {
			approximation = rl
			break
		}
	}
	fmt.Println("got:", got, " approximation:", strconv.FormatFloat(approximation, 'f', -1, 64))
}

func printFloat(f float64) {
	fmt.Println(strconv.FormatFloat(f, 'f', -1, 64))
	fmt.Println()
}

func Ok(xxx string, ok bool, args ...interface{}) {
	var s string
	if ok {
		s = " is ok"
	} else {
		s = " is not ok"
	}
	fmt.Println(fmt.Sprintf(xxx+s, args...))
	fmt.Println()
}

func getNum(base, k, b, f int, cnt *int) {
	for i := base; i < 1<<f; i++ {
		if k+1 < b {
			getNum(i, k+1, b, f, cnt)
		} else {
			*cnt++
		}
	}
}

func getNextPow2(n uint64) uint {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return uint(n)
}

func getNumOfKindAndBit(b, f int) {
	cnt := 0
	getNum(0, 0, b, f, &cnt)
	fmt.Printf("Num of kinds: %v, Num of needed bits: %v\n", cnt, math.Log2(float64(getNextPow2(uint64(cnt)))))
}

// GetLoadFactor 获得负载因子
func GetLoadFactor(b uint32) float64 {
	switch b {
	case 2:
		return 0.84
	case 4:
		return 0.95
	case 8:
		return DefaultLoadFactor
	default:
		return DefaultLoadFactor

	}
}

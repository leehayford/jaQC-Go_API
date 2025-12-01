package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/stat" 		// go get gonum.org/v1/gonum/...
	"github.com/google/uuid"     	// go get github.com/google/uuid
)

type XYPoint struct {
	X int64   `json:"x"`
	Y float32 `json:"y"`
}

func ValidateUUIDString(u string) (ok bool) {
	if u == "" || u == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	_, err := uuid.Parse(u)
	return err == nil
}

func StructToJSONString(obj interface{}) (str string, err error) {
	js, err := json.Marshal(obj)
	if err != nil {
		return
	}
	str = string(js)
	return
}

func JSONStringToStruct(js string, obj interface{}) (err error) {
	if err = json.Unmarshal([]byte(js), &obj); err != nil {
		err = fmt.Errorf("error converting json string to struct: %s", err.Error())
	}
	return
}

/*BYTES OUTPUT*/
func GetBytes_B(v any) []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, v)
	if err != nil {
		fmt.Println(err)
	}
	return buffer.Bytes()
}
func GetBytes_L(v any) []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.LittleEndian, v)
	if err != nil {
		fmt.Println(err)
	}
	return buffer.Bytes()
}

/*BYTES INPUT*/
func BytesToUInt16_B(bytes []byte) uint16 {
	x := make([]byte, 2)
	i := len(x) - len(bytes)
	n := len(bytes) - 1
	// fmt.Println("Received bytes:\t", n)
	for n >= 0 {
		x[n+i] = bytes[n]
		n--
	}
	// fmt.Printf("Final bytes:\t%d\t%x\n", len(x), x)
	return binary.BigEndian.Uint16(x)
}
func BytesToUInt16_L(bytes []byte) uint16 {
	x := make([]byte, 2) 	// fmt.Println("Received bytes:\t", len(bytes))
	copy(x, bytes) 			// fmt.Printf("Final bytes:\t%d\t%x\n", len(x), x)
	return binary.LittleEndian.Uint16(x)
}

func BytesToUInt32_B(bytes []byte) uint32 {
	x := make([]byte, 4)
	i := len(x) - len(bytes)
	n := len(bytes) - 1
	// fmt.Println("Received bytes:\t", len(bytes))
	for n >= 0 {
		x[n+i] = bytes[n]
		n--
	}
	// fmt.Printf("Final bytes:\t%d\t%x\n", len(x), x)
	return binary.BigEndian.Uint32(x)
}
func BytesToUInt32_L(bytes []byte) uint32 {
	x := make([]byte, 4) 	// fmt.Println("Received bytes:\t", len(bytes))
	copy(x, bytes) 			// fmt.Printf("Final bytes:\t%d\t%x\n", len(x), x)
	return binary.LittleEndian.Uint32(x)
}
func BytesToInt32_L(bytes []byte) int32 {
	return int32(BytesToUInt32_L(bytes))
}

func BytesToInt64_B(bytes []byte) int64 {
	return int64(binary.BigEndian.Uint64(bytes))
}
func BytesToInt64_L(bytes []byte) int64 {
	return int64(binary.LittleEndian.Uint64(bytes))
}
func BytesToUint64_L(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}

func BytesToFloat32_B(bytes []byte) float32 {
	return float32(BytesToUInt32_B(bytes))
	// return math.Float32frombits(BytesToUInt32_B(bytes))
}
func BytesToFloat32_L(bytes []byte) float32 {
	return math.Float32frombits(BytesToUInt32_L(bytes))
}
func BytesToFloat64_L(bytes []byte) float64 {
	return math.Float64frombits(BytesToUint64_L(bytes))
}

func BytesToBase64(bytes []byte) string {
	str := base64.StdEncoding.EncodeToString(bytes)
	return str
}
func Base64ToBytes(b64 string) []byte {
	bytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		fmt.Println(err)
	}
	return bytes
}

func BytesToBase64URL(bytes []byte) string {
	str := base64.URLEncoding.WithPadding(-1).EncodeToString(bytes)
	return str
}
func Base64URLToBytes(b64 string) (bytes []byte, err error) {
	bytes, err = base64.URLEncoding.WithPadding(-1).DecodeString(b64)
	if err != nil {
		fmt.Println(err)
	}
	return bytes, err
}

func Int64ToBytes(in int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(in))
	return b
}

func Int32ToBytes(in int32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(in))
	return b
}

func Int16ToBytes(in int16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(in))
	return b
}

func Float32ToBytes(in float32) []byte {

	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, in); err != nil {
		LogErr(err)
	}
	return b.Bytes()
}

func Float64ToBytes(in float64) []byte {

	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, in); err != nil {
		LogErr(err)
	}
	return b.Bytes()
}

// func StrBytesToString(b []byte) (out string) {
// 	for i := range b {
// 		if b[i] != 255 {
// 			return string(b[i:])
// 		}
// 	}
// 	return
// }

func StrBytesToString(b []byte) string {
	for i := range b {
		if b[i] == 32 || b[i] == 255 {
			return string(b[:i])
		}
	}
	return string(b)
}

/* STRING INPUT */
func StringToNBytes(str string, size int) []byte {

	bin := []byte(str)
	l := len(bin)

	if l == size {
		/* bin ALREADY THE RIGHT SIZE, SHIP IT */
		return bin
	}
	if l > size {
		/* bin TOO BIG, RETURN THE LAST 'size' BYTES
		WE COULD RETURN THE FIRST 'size' BYTES...
		*/
		return bin[l-size:]
	}

	/* bin TOO SMALL*/

	/* FILL BUFFER WITH 'size' SPACES */
	out := bytes.Repeat([]byte{0x20}, size)

	/* WRITE 'bin TO THE START OF THE BUFFER */
	copy(out[:l], bin)

	// fmt.Printf("\n%s ( %d ) : %x\n",str , len(out), out)
	return out
}
func ValidateStringLength(str string, size int) (out string) {

	if len(str) > size {
		/* str TOO BIG, RETURN THE  FIRST 'size' CHARS... */
		return str[:size]
	}
	/* str ALREADY THE RIGHT SIZE, SHIP IT */
	return str
}

// func Float32ToHex(f float32)

func StringToInt64(str string) int64 {
	out, err := strconv.ParseInt(strings.Trim(str, " "), 0, 64)
	if err != nil {
		pc, file, line, _ := runtime.Caller(1)
		name := runtime.FuncForPC(pc).Name()
		fmt.Printf("***ERROR***\nFile:\t%s\nFunc  :\t%s\nLine  :\t%d\nError :\n%s", file, name, line, err.Error())
		return 0
	}
	return out
}
func StringToInt32(str string) int32 {
	return int32(StringToInt64(str))
}

func StringToFloat64(str string) float64 {
	out, err := strconv.ParseFloat(strings.Trim(str, " "), 32)
	if err != nil {
		pc, file, line, _ := runtime.Caller(1)
		name := runtime.FuncForPC(pc).Name()
		fmt.Printf("***ERROR***\nFile:\t%s\nFunc  :\t%s\nLine  :\t%d\nError :\n%s", file, name, line, err.Error())
		return 0
	}
	return out
}
func StringToFloat32(str string) float32 {
	return float32(StringToFloat64(str))
}

func MinMaxUInt32(slice []uint32) (uint32, uint32) {

	min := slice[0]
	max := slice[0]
	for _, v := range slice {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	return min, max
}

func MinMaxFloat32(slice []float32, margin float32) (float32, float32) {

	min := slice[0]
	max := slice[0]
	for _, v := range slice {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	span := max - min
	min -= span * margin
	max += span * margin
	// fmt.Printf("MIN: %f, MAX: %f\n", min, max)
	return min, max
}

func MeanInt32(slice []int32) int32 {
	var mean int32
	for _, val := range slice {
		mean += val
	}
	mean = mean / int32(len(slice))
	return mean
}
func MeanFloat32(slice []float32) float32 {
	var mean float32
	for _, val := range slice {
		mean += val
	}
	mean = mean / float32(len(slice))
	return mean
}


func SlopeInterceptDeviation(xMean, yMean float32, xs, ys []float32) (m, b, devi float32) {
	
	SumXYProd := float32(0)

	SumSqXs := float32(0)

	SumSqYs := float32(0)

	for i := 0; i < len(xs); i++ {

		SumXYProd += (xs[i] - xMean) * (ys[i] - yMean)

		SumSqXs += (xs[i] - xMean) * (xs[i] - xMean)

		SumSqYs += (ys[i] - yMean) * (ys[i] - yMean)
	}
	
	m = SumXYProd  /  SumSqXs

	b = yMean - (m * xMean)

	devi = float32(math.Sqrt(float64(SumSqYs)/float64(len(ys))))

	return 
}

func SlopeAndIntercept(xMean, yMean float32, xs, ys []float32) (m, b float32) {
	
	SumXYProd := float32(0)

	SumSqXs := float32(0)

	for i := 0; i < len(xs); i++ {

		SumXYProd += (xs[i] - xMean) * (ys[i] - yMean)

		SumSqXs += (xs[i] - xMean) * (xs[i] - xMean)
	}
	
	m = SumXYProd  /  SumSqXs

	b = yMean - (m * xMean)

	return 
}

func SlopeAndIntercept_XInt32(xs []int32, ys []float32) (m float32, b float32) {
	
	SumXYProd := float32(0)

	SumSqXs := float32(0)

	xMean := float32(MeanInt32(xs))

	yMean:= MeanFloat32(ys)

	for i := 0; i < len(xs); i++ {

		SumXYProd += (float32(xs[i]) - xMean) * (ys[i] - yMean)

		SumSqXs += (float32(xs[i]) - xMean) * (float32(xs[i]) - xMean)
	}
	
	m = SumXYProd  /  SumSqXs

	b = yMean - (m * xMean)

	return 
}


func MeanStdDev(yArr []float32) (mean, std float64) {
	var arr []float64

	for _, v := range yArr {
		arr = append(arr, float64(v))
	}

	return stat.MeanStdDev(arr, nil)
}

func MeanStdDev_F32(yArr []float32) (mean, std float32) {
	m, s := MeanStdDev(yArr)
	mean = float32(m)
	std = float32(s) 
	return 
}

func StandardDeviation(ySlice []float64) (devi float64) {
	_, devi = stat.MeanStdDev(ySlice, nil)
	return devi
}

func StandardDeviation_F32(ySlice []float64) (devi float32) {
	return float32(StandardDeviation(ySlice))
}

type TSXY struct {
	X []int64
	Y []float32
}

func (v TSXY) TSXs() []int64 {
	return v.X
}
func (v TSXY) TSYs() []float32 {
	return v.Y
}
func (v TSXY) MinMax(margin float32) (float32, float32) {
	return MinMaxFloat32(v.TSYs(), margin)
}
func (v TSXY) TSD(margin float32) TimeSeriesData {
	min, max := v.MinMax(margin)
	return TimeSeriesData{TSDPoints(v), min, max}
}

type TSValues interface {
	TSXs() []int64
	TSYs() []float32
}
type TSDPoint struct {
	X int64   `json:"x"`
	Y float32 `json:"y"`
}

func TSDPoints(v TSValues) []TSDPoint {
	xs, ys := v.TSXs(), v.TSYs()
	points := []TSDPoint{}
	for i, x := range xs {
		point := TSDPoint{}
		point.X = x
		point.Y = ys[i]
		points = append(points, point)
	}

	return points
}

type TimeSeriesData struct {
	Data []TSDPoint `json:"data"`
	Min  float32    `json:"min"`
	Max  float32    `json:"max"`
}

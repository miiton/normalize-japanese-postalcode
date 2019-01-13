package main

import (
	"io"
	"os"
	"strings"

	"encoding/csv"

	"github.com/miiton/kanaconv"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

/*
1. CSVをSJIS->UTF8に変換
2. 9カラム目 "以下に掲載がない場合" を "" に置換
3. 9カラム目 "境町の次に番地がくる場合" を "" に置換
4. 9カラム目 ".*[村町]一円" を "" に置換 ※ "一円" から始まる場合は存在する地名にあたるので置換しない
5. 9カラム目が長すぎて分割されているパターンがあるのでマージする(例: 6028119)
6. 半角カタカナは全角カタカナに置換する
*/

// KenAll 郵便番号データ構造体
type KenAll struct {
	// 全国地方公共団体コード（JIS X0401、X0402）………　半角数字
	JISX0402 string
	// （旧）郵便番号（5桁）………………………………………　半角数字
	OldPostal string
	// 郵便番号（7桁）………………………………………　半角数字
	Postal string
	// 都道府県名　…………　半角カタカナ（コード順に掲載）　（注1）
	Level1Kana string
	// 市区町村名　…………　半角カタカナ（コード順に掲載）　（注1）
	Level2Kana string
	// 町域名　………………　半角カタカナ（五十音順に掲載）　（注1）
	Line1Kana string
	// 都道府県名　…………　漢字（コード順に掲載）　（注1,2）
	Level1 string
	// 市区町村名　…………　漢字（コード順に掲載）　（注1,2）
	Level2 string
	// 町域名　………………　漢字（五十音順に掲載）　（注1,2）
	Line1 string
	// 一町域が二以上の郵便番号で表される場合の表示　（注3）　（「1」は該当、「0」は該当せず）
	Option1 string
	// 小字毎に番地が起番されている町域の表示　（注4）　（「1」は該当、「0」は該当せず）
	Option2 string
	// 丁目を有する町域の場合の表示　（「1」は該当、「0」は該当せず）
	Option3 string
	// 一つの郵便番号で二以上の町域を表す場合の表示　（注5）　（「1」は該当、「0」は該当せず）
	Option4 string
	// 更新の表示（注6）（「0」は変更なし、「1」は変更あり、「2」廃止（廃止データのみ使用））
	Option5 string
	// 変更理由　（「0」は変更なし、「1」市政・区政・町政・分区・政令指定都市施行、「2」住居表示の実施、「3」区画整理、「4」郵便区調整等、「5」訂正、「6」廃止（廃止データのみ使用））
	Option6 string
}

// Jigyosyo 事業所の個別郵便番号
type Jigyosyo struct {
	// 大口事業所の所在地のJISコード（5バイト）
	JISX0402 string
	// 大口事業所名（カナ）（100バイト）
	Kana string
	// 大口事業所名（漢字）（160バイト）
	Name string
	// 都道府県名（漢字）（8バイト）
	Level1 string
	// 市区町村名（漢字）（24バイト）
	Level2 string
	// 町域名（漢字）（24バイト）
	Line1 string
	// 小字名、丁目、番地等（漢字）（124バイト）
	Line2 string
	// 大口事業所個別番号（7バイト）
	Postal string
	// 旧郵便番号（5バイト）
	OldPostal string
	// 取扱局（漢字）（40バイト）
	Option7 string
	// 個別番号の種別の表示（1バイト）
	//     「0」大口事業所
	//     「1」私書箱
	Option8 string
	// 複数番号の有無（1バイト） 「0」複数番号無し
	//     「1」複数番号を設定している場合の個別番号の1
	//     「2」複数番号を設定している場合の個別番号の2
	//     「3」複数番号を設定している場合の個別番号の3
	Option9 string
	// 修正コード（1バイト）
	//     「0」修正なし
	//     「1」新規追加
	//     「5」廃止
	Option10 string
}

func replaceLine1(s string) string {
	if s == "以下に掲載がない場合" {
		return ""
	}
	if s == "境町の次に番地がくる場合" {
		return ""
	}
	if strings.HasPrefix(s, "一円") {
		return s
	}
	if strings.HasSuffix(s, "一円") {
		return ""
	}
	return s
}

func unmarshalKenAll(record []string) KenAll {
	postal := KenAll{
		JISX0402:   record[0],
		OldPostal:  record[1],
		Postal:     record[2],
		Level1Kana: kanaconv.SmartConv(record[3]),
		Level2Kana: kanaconv.SmartConv(record[4]),
		Line1Kana:  kanaconv.SmartConv(record[5]),
		Level1:     kanaconv.SmartConv(record[6]),
		Level2:     kanaconv.SmartConv(record[7]),
		Line1:      replaceLine1(kanaconv.SmartConv(record[8])),
		Option1:    record[9],
		Option2:    record[10],
		Option3:    record[11],
		Option4:    record[12],
		Option5:    record[13],
		Option6:    record[14],
	}
	return postal
}

func mergeLine1(bufPostal []KenAll) []KenAll {
	var level3 string
	for _, p := range bufPostal {
		level3 = level3 + p.Line1
	}
	postal := bufPostal[0]
	postal.Line1 = level3
	return []KenAll{postal}

}

func parseKenAll(csvFile *os.File, writer *csv.Writer) {
	reader := csv.NewReader(transform.NewReader(csvFile, japanese.ShiftJIS.NewDecoder()))
	bufPostal := []KenAll{}
	lastPostal := KenAll{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		p := unmarshalKenAll(record)

		// 1つ前の郵便番号と異なる場合はバッファを初期化
		if lastPostal.Postal != p.Postal {
			if strings.HasSuffix(lastPostal.Line1, ")") && !strings.Contains(lastPostal.Line1, "(") && len(bufPostal) > 1 {
				bufPostal = mergeLine1(bufPostal)
			}
			for _, b := range bufPostal {
				row := []string{
					b.JISX0402,
					b.OldPostal,
					b.Postal,
					b.Level1Kana,
					b.Level2Kana,
					b.Line1Kana,
					b.Level1,
					b.Level2,
					b.Line1,
					"",
					"",
					"",
					b.Option1,
					b.Option2,
					b.Option3,
					b.Option4,
					b.Option5,
					b.Option6,
					"",
					"",
					"",
					"",
				}
				writer.Write(row)
			}
			bufPostal = []KenAll{}
		}
		bufPostal = append(bufPostal, p)
		lastPostal = p
	}
}

func unmarshalJigyosyo(record []string) Jigyosyo {
	postal := Jigyosyo{
		JISX0402:  record[0],
		Kana:      kanaconv.SmartConv(record[1]),
		Name:      kanaconv.SmartConv(record[2]),
		Level1:    kanaconv.SmartConv(record[3]),
		Level2:    kanaconv.SmartConv(record[4]),
		Line1:     kanaconv.SmartConv(record[5]),
		Line2:     kanaconv.SmartConv(record[6]),
		Postal:    record[7],
		OldPostal: record[8],
		Option7:   record[9],
		Option8:   record[10],
		Option9:   record[11],
		Option10:  record[12],
	}
	return postal
}

func parseJigyosyo(csvFile *os.File, writer *csv.Writer) {
	reader := csv.NewReader(transform.NewReader(csvFile, japanese.ShiftJIS.NewDecoder()))
	bufPostal := []Jigyosyo{}
	lastPostal := Jigyosyo{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		p := unmarshalJigyosyo(record)

		// 1つ前の郵便番号と異なる場合はバッファを初期化
		if lastPostal.Postal != p.Postal {
			for _, b := range bufPostal {
				row := []string{
					b.JISX0402,
					b.OldPostal,
					b.Postal,
					"",
					"",
					"",
					b.Level1,
					b.Level2,
					b.Line1,
					b.Line2,
					b.Name,
					b.Kana,
					"",
					"",
					"",
					"",
					"",
					"",
					b.Option7,
					b.Option8,
					b.Option9,
					b.Option10,
				}
				writer.Write(row)
			}
			bufPostal = []Jigyosyo{}
		}
		bufPostal = append(bufPostal, p)
		lastPostal = p
	}
}

func main() {
	outputCsv, err := os.Create("./postal.csv")
	if err != nil {
		panic(err)
	}
	defer outputCsv.Close()

	writer := csv.NewWriter(outputCsv)
	if err != nil {
		panic(err)
	}
	defer writer.Flush()

	kenAll, err := os.Open("./KEN_ALL.CSV")
	if err != nil {
		panic(err)
	}
	defer kenAll.Close()

	parseKenAll(kenAll, writer)

	jigyosyo, err := os.Open("./JIGYOSYO.CSV")
	if err != nil {
		panic(err)
	}
	defer jigyosyo.Close()

	parseJigyosyo(jigyosyo, writer)
}

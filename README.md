# normalization-ken_all
郵便番号データをDBで扱いやすく変換する

## データの準備

```sh
curl -LO https://www.post.japanpost.jp/zipcode/dl/kogaki/zip/ken_all.zip
unzip ken_all.zip
curl -LO "http://www.post.japanpost.jp/zipcode/dl/jigyosyo/zip/jigyosyo.zip" -o jigyosyo.zip
unzip jigyosyo.zip
```

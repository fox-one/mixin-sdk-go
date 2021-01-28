module github.com/fox-one/mixin-sdk-go

go 1.15

require (
	github.com/btcsuite/btcutil v1.0.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-resty/resty/v2 v2.3.0
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/gorilla/websocket v1.4.2
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c
	github.com/shopspring/decimal v1.2.0
	github.com/stretchr/testify v1.6.1
	github.com/vmihailenco/msgpack/v4 v4.0.0+incompatible
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	google.golang.org/appengine v1.6.7 // indirect
)

replace github.com/vmihailenco/msgpack/v4 => github.com/MixinNetwork/msgpack/v4 v4.3.13

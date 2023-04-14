module go-captcha-example

go 1.18

require (
	github.com/allegro/bigcache/v3 v3.1.0
	gitlab.test.com/common/log v1.1.3
	gitlab.test.com/video-platform/common v1.0.4-0.20230411083841-9b1507ceff22
)

require (
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible // indirect
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/image v0.6.0 // indirect
)

replace gitlab.test.com/video-platform/common v1.0.4-0.20230411083841-9b1507ceff22 => ../common

module github.com/codenotary/immugorm

go 1.13

require (
	github.com/codenotary/immudb v1.2.2-0.20211224171643-06d4378fbf62
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.40.0
	gorm.io/gorm v1.22.4
)

replace github.com/spf13/afero => github.com/spf13/afero v1.5.1

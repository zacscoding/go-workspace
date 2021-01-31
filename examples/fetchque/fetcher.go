package fetchque

type (
	Data1Callback func(data Data1)
	Data2Callback func(data Data2)
)

type Fetcher interface {
	FetchData1(key string, callback Data1Callback) error

	FetchData2(key string, callback Data2Callback) error
}

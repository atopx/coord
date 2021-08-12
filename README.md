# coord
坐标相关工具

# download
```
go get github.com/yanmengfei/coord@v0.0.1
```

# 使用
```go
package main

import (
    "fmt"
    "github.com/yanmengfei/coord"
)

func main() {
    var location = "115.668055,34.449162"
    var lon, lat, _ = coord.LocationToFloat64Coord(location)
    fmt.Println("GCJ02:", lon, lat)
    lon, lat = coord.GCJ02toWGS84(lon, lat)
    fmt.Println("WGS84:", lon, lat)
}

```


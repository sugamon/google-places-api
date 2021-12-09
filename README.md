# google-places-api
google-places-api is a [Google Places API](https://developers.google.com/maps/documentation/places/web-service/search) wrap in Go

# Installation
```
go get github.com/sugamon/google-places-api
```

# Provides API
- TextSearch

# Usage
## TextSearch
```
import (
	"context"
	places "github.com/sugamon/google-places-api"
)

apiKey := os.Getenv("API_KEY")
p := places.NewClient(apiKey)

optionalParams := places.OptionalParams{
  Language: "en",
  Location: "uk",
}

res, err := p.TextSearch(context.Background(), "london beer", optionalParams)
```
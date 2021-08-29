# Good With Buffalo

This is the source code for [withbuffalo.com](https://withbuffalo.com/)

## Development mode

```DRY_RUN=1 go run main.go```

## Production mode

```bash
docker build -t buffalo .
docker run -d --restart always --name buffalo \
        -p 8000:8000 \
        -e MYSQL_DB=username:password@protocol(address)/dbname \
        -e API_KEY=XXXXXXX \
        -e API_SECRET=YYYYYYYY \
        -e ACCESS_TOKEN=XYXYXYXYX \
        -e ACCESS_TOKEN_SECRET=YXYXYXYXYX \
        buffalo
```

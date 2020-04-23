docker build -t yuruh/encrypted-diary:$(git describe --abbrev=0) .

docker push yuruh/encrypted-diary:$(git describe --abbrev=0)
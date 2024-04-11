echo test > test.txt
curl -X POST -F "file=@test.txt" http://localhost:18080/upload
curl -O http://localhost:18080/download/test.txt

curl -u username:password -X POST -F "file=@test.txt" http://127.0.0.1:18080/upload
curl -u username:password -O http://127.0.0.1:18080/download/test.txt

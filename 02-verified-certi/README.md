openssl genpkey -algorithm RSA -out private.pem
openssl rsa -pubout -in private.pem -out public.pem
openssl dgst -sha256 -sign private.pem -out signature_file example.exe
go run main.go public.pem executable_file signature_file
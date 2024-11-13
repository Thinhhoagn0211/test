rm *.pem
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=VN/ST=HCM/L=HCM/O=Company/OU=Education/CN=thinh/emailAddress=raven4work0211@gmail.com"
echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=VN/ST=HN/L=HN/O=Company/OU=Education/CN=thanh/emailAddress=raven4work0211@gmail.com"
openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.conf
echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text
openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C=VN/ST=BD/L=Company/O=Checker/OU=Computer/CN=thai/emailAddress=raven4work0211@gmail.com"
openssl x509 -req -in client-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile client-ext.conf
echo "Client's signed certificate"
openssl x509 -in client-cert.pem -noout -text
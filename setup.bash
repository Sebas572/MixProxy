# Clonar el repositorio
git clone https://github.com/Sebas572/MixProxy.git
cd MixProxy

wget https://github.com/Sebas572/MixProxy/releases/download/dev/mixproxy-linux-amd64

rm -r .config

./mixproxy-linux-amd64

echo ""
echo "Please place your SSL certificates in the mixproxy/certs folder. Alternatively, if you are in development mode, you can run mixproxy and select 'Create certificates SSL (developer)'."
echo ""
echo "After configuring the certificates, run docker compose up --build."
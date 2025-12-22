# Clonar el repositorio
git clone https://github.com/Sebas572/MixProxy.git --depth=1
cd MixProxy

Invoke-WebRequest -Uri "https://github.com/Sebas572/MixProxy/releases/download/v0.0.2-experimental/mixproxy.exe" -OutFile "mixproxy.exe"

rm -r .config

./mixproxy

Write-Host ""
Write-Host "Please place your SSL certificates in the mixproxy/certs folder. Alternatively, if you are in development mode, you can run mixproxy and select 'Create certificates SSL (developer)'."
Write-Host ""
Write-Host "After configuring the certificates, run docker compose up --build."
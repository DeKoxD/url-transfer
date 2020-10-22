# URL Transfer

Service made to help send web sites from any browser to any browser (e.g. Kindle, PC, phone).

## How to use

1. Run on terminal:
```shell
sudo docker-compose up --build
```
2. Open _pc-hostname_/receive on the device you want to receive an URL. An QR code will show up.
3. On the device that has the URL you want to send:
* Read the QR code, or;
* Go to _pc-hostname_/send and type the ID above the QR code.
4. Paste the QR Code you want to send on the text box and press _Send_.
5. Press _Get URL_ on the device you want to receive.
ADDRESS = localhost:3000

tests: get_test form_test json_test

run:
	go get ./...
	go run main.go

get_test:
	curl -X GET 'http://localhost:3000/locations?url=https://helpx.adobe.com/in/stock/how-to/visual-reverse-image-search/_jcr_content/main-pars/image.img.jpg/visual-reverse-image-search-v2_1000x560.jpg'

form_test:
	curl -F 'image=@test_data/for_form.png' $(ADDRESS)

json_test:
	(echo -n '{"image": "'; base64 test_data/json.jpg; echo '"}') | \
	curl -H "Content-Type: application/json" -d @-  $(ADDRESS) 

clean:
	rm -r thumbnails










ifeq ("x","y")

sudo apt install build-essential pkg-config glib2.0-dev libexpat1-dev

cd /tmp/

# This assumed a UNIX based OS. Download the appropriate version from https://github.com/jcupitt/libvips/releases
# if you are running Windows
wget https://github.com/jcupitt/libvips/releases/download/v8.6.2/vips-8.6.2.tar.gz

tar xzf vips-8.6.2.tar.gz
cd vips-8.6.2
./configure
make
sudo make install
sudo ldconfig

endif
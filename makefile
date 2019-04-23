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
FROM golang:1.12


RUN apt-get update \
    && apt-get -y upgrade \
    && apt-get install -y pkg-config \
    && apt-get install -y wget \
    && apt-get install -y make \
    && apt-get install -y gcc \
    && apt-get install -y gtk-doc-tools \
    && apt-get install -y libtool \
    && apt-get install -y autoconf \ 
    && apt-get install -y gobject-introspection \
    && apt-get install -y git


RUN apt-get install -y \
    build-essential \
    glib-2.0-dev \
    libexpat-dev \
    libexpat1-dev \
    librsvg2-dev \
    libpng-dev \
    libgif-dev \
    libjpeg-dev \
    libexif-dev \
    liblcms2-dev \
    liborc-dev 

WORKDIR /tmp/
RUN apt-get -y install build-essential pkg-config glib2.0-dev libexpat1-dev \
    && wget https://github.com/jcupitt/libvips/archive/v8.7.0.tar.gz

RUN tar xzf v8.7.0.tar.gz \
    && cd libvips-8.7.0 \
    #&& chmod +x ./configure.ac \
    #&& ./configure \
    && bash autogen.sh \
    && make \
    && make install \
    && ldconfig


WORKDIR /go/src

RUN git clone https://github.com/zaur22/staply_img_resizer.git \
    && cd staply_img_resizer \
    #&& go get -u github.com/davidbyttow/govips/pkg/vips \
    && go get  ./... \
    && go install -v ./... \
    && go build main.go \
    && ./main.go
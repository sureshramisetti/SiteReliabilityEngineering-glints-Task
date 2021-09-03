FROM ubuntu:18.04

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends \
        autoconf \
        automake \
        bison \
        build-essential \
        ca-certificates \
        flex \
        gawk \
        gettext \
        gperf \
        libtool \
        pkg-config \
        sudo \
        zlib1g-dev \
        libgmp3-dev \
        libmpfr-dev \
        libmpc-dev \
        texinfo \
        git \
        vim && \
    rm -rf /var/lib/apt/lists

RUN cd /opt && \
    git clone --depth=1 https://bitbucket.org/padavan/rt-n56u.git

RUN cd /opt/rt-n56u/toolchain-mipsel && \
    ./clean_sources && \
    ./build_toolchain

WORKDIR /opt/rt-n56u/trunk

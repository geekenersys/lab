FROM ubuntu:24.04

# Expose necessary ports
EXPOSE 5900
EXPOSE 5901
EXPOSE 8080

# Environment variable for VNC password
ENV VNC_PASSWD=123456

# Install necessary packages
ENV APT_INSTALL_PRE="apt -o Acquire::ForceIPv4=true update && DEBIAN_FRONTEND=noninteractive apt -o Acquire::ForceIPv4=true install -y --no-install-recommends"
ENV APT_INSTALL_POST="&& apt clean -y && rm -rf /var/lib/apt/lists/*"
RUN eval ${APT_INSTALL_PRE} tigervnc-standalone-server tigervnc-common tigervnc-tools fluxbox eterm xterm git net-tools python3 python3-numpy ca-certificates scrot libx11-dev libxtst-dev libxinerama-dev libxrandr-dev libxcursor-dev libxcb1-dev libx11-xcb-dev libxkbcommon-dev libxkbcommon-x11-dev libxkbfile-dev gcc clang cmake make chromium-browser libxt-dev ${APT_INSTALL_POST}

# Install VNC and websockify
RUN git clone --branch v1.4.0 --single-branch https://github.com/novnc/noVNC.git /opt/noVNC
RUN git clone --branch v0.11.0 --single-branch https://github.com/novnc/websockify.git /opt/noVNC/utils/websockify
RUN ln -s /opt/noVNC/vnc.html /opt/noVNC/index.html

# Install Go
RUN apt-get update && apt-get install -y golang-go
RUN go version

# Add menu entries
RUN echo "?package(bash):needs=\"X11\" section=\"DockerCustom\" title=\"Xterm\" command=\"xterm -ls -bg black -fg white\"" >> /usr/share/menu/custom-docker && update-menus

# Set timezone to UTC
RUN ln -snf /usr/share/zoneinfo/UTC /etc/localtime && echo UTC > /etc/timezone

# Health check
HEALTHCHECK --start-period=10s CMD bash -c "if [ \"`pidof -x Xtigervnc | wc -l`\" == '1' ]; then exit 0; else exit 1; fi"

# Add non-root user
ENV UID_OF_DOCKERUSER 1001
RUN useradd -m -s /bin/bash -g users -u ${UID_OF_DOCKERUSER} dockerUser
RUN chown -R dockerUser:users /home/dockerUser && chown dockerUser:users /opt

# Create shared directory
RUN mkdir -p /shared && chmod -R 777 /shared

# Switch to non-root user
USER dockerUser

# Copy scripts and Go files
COPY --chown=dockerUser:users container_startup.sh /opt/container_startup.sh
COPY --chown=dockerUser:users x11vnc_entrypoint.sh /opt/x11vnc_entrypoint.sh
COPY --chown=dockerUser:users main.go /home/dockerUser/main.go

# Initialize Go modules
WORKDIR /home/dockerUser
RUN go mod init robovncx
RUN go get github.com/go-vgo/robotgo
RUN go get github.com/robotn/gohook

# Build and install libuiohook
RUN git clone https://github.com/kwhat/libuiohook /home/dockerUser/libuiohook
WORKDIR /home/dockerUser/libuiohook
RUN mkdir build && cd build && cmake -S .. -D BUILD_SHARED_LIBS=ON -D BUILD_DEMO=ON -DCMAKE_INSTALL_PREFIX=../dist && cmake --build . --parallel 2 --target install

# Set DISPLAY environment variable
ENV DISPLAY=:0

# Build the Go application
WORKDIR /home/dockerUser
RUN go build -o app main.go

# Entry point
ENTRYPOINT ["/opt/container_startup.sh"]

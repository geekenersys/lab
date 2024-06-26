#!/bin/bash
OUR_IP=$(hostname -i)

# start VNC server (Uses VNC_PASSWD Docker ENV variable)
mkdir -p $HOME/.vnc && echo "$VNC_PASSWD" | vncpasswd -f > $HOME/.vnc/passwd
# Remove potential lock files created from a previously stopped session
rm -rf /tmp/.X*
echo "Starting VNC server"
vncserver :0 -localhost no -nolisten -rfbauth $HOME/.vnc/passwd -xstartup /opt/x11vnc_entrypoint.sh &

echo "Starting noVNC web server"
/opt/noVNC/utils/novnc_proxy --vnc localhost:5900 --listen 5901 &

echo "Starting fluxbox window manager"
/usr/bin/fluxbox &

echo "Starting Golang HTTP server"
/home/dockerUser/app &

echo -e "\n\n------------------ VNC environment started ------------------"
echo -e "\nVNCSERVER started on DISPLAY= $DISPLAY \n\t=> connect via VNC viewer with $OUR_IP:5900"
echo -e "\nnoVNC HTML client started:\n\t=> connect via http://$OUR_IP:5901/?password=$VNC_PASSWD\n"
echo -e "\nGolang HTTP server started:\n\t=> connect via http://$OUR_IP:8081\n"

if [ -z "$1" ]; then
  tail -f /dev/null
else
  # unknown option ==> call command
  echo -e "\n\n------------------ EXECUTE COMMAND ------------------"
  echo "Executing command: '$@'"
  exec $@
fi

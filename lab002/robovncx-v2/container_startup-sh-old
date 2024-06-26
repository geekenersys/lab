#!/bin/bash
OUR_IP=$(hostname -i)

# Start VNC server (Uses VNC_PASSWD Docker ENV variable)
mkdir -p $HOME/.vnc
echo "$VNC_PASSWD" | vncpasswd -f > $HOME/.vnc/passwd

# Setup X11 authentication
touch $HOME/.Xauthority
xauth generate :0 . trusted
xauth add :0 . $(mcookie)

# Remove potential lock files created from a previously stopped session
rm -rf /tmp/.X*
rm -rf /tmp/.X11-unix
rm -rf /tmp/.X*-lock

echo "Starting VNC server"
vncserver -kill :0 &> /dev/null || echo "No existing VNC server to kill"
vncserver :0 -localhost no -nolisten tcp -rfbauth $HOME/.vnc/passwd -xstartup /opt/x11vnc_entrypoint.sh &
VNCSERVER_PID=$!
sleep 5

# Check if the VNC server process is running
if ps -p $VNCSERVER_PID > /dev/null; then
   echo "VNC server is running"
else
   echo "VNC server failed to start, retrying..."
   vncserver -kill :0 &> /dev/null || echo "No existing VNC server to kill"
   vncserver :0 -localhost no -nolisten tcp -rfbauth $HOME/.vnc/passwd -xstartup /opt/x11vnc_entrypoint.sh &
   VNCSERVER_PID=$!
   sleep 5
   if ps -p $VNCSERVER_PID > /dev/null; then
      echo "VNC server is running after retry"
   else
      echo "VNC server failed to start after retry"
      tail -n 100 $HOME/.vnc/*.log
      exit 1
   fi
fi

echo "Starting noVNC web server"
/opt/noVNC/utils/novnc_proxy --vnc localhost:5900 --listen 5901 &
NOVNC_PID=$!
sleep 5

echo "Starting fluxbox window manager"
/usr/bin/fluxbox &
FLUXBOX_PID=$!
sleep 5

echo "Starting Golang HTTP server"
/home/dockerUser/app &
APP_PID=$!
sleep 5

# Check if the processes are running
echo "Checking if noVNC server is running"
if ps -p $NOVNC_PID > /dev/null; then
   echo "noVNC server is running"
else
   echo "noVNC server failed to start"
   exit 1
fi

echo "Checking if fluxbox is running"
if ps -p $FLUXBOX_PID > /dev/null; then
   echo "fluxbox is running"
else
   echo "fluxbox failed to start"
   exit 1
fi

echo "Checking if Golang HTTP server is running"
if ps -p $APP_PID > /dev/null; then
   echo "Golang HTTP server is running"
else
   echo "Golang HTTP server failed to start"
   exit 1
fi

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

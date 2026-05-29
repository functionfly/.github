#!/bin/sh

# This script configures a local syslog daemon to forward its logs to SWO.
# Download the latest version from https://agent-binaries.cloud.solarwinds.com
#
# Usage: SWO_HOSTNAME=... SWO_TOKEN=... sudo -E solarwinds-syslog-setup.sh
# The Add Data wizard provides customized instructions for your account.

# The script will prompt for SWO_HOSTNAME and SWO_TOKEN env vars if necessary.
SWO_HOSTNAME=${SWO_HOSTNAME:-}
SWO_PORT=${SWO_PORT:-6514}
SWO_TOKEN=${SWO_TOKEN:-}

VERSION="20221012"

RSYSLOG_DIR=${RSYSLOG_DIR:-"/etc/rsyslog.d"}
SYSLOGNG_DIR=${SYSLOGNG_DIR:-"/etc/syslog-ng/conf.d"}
SYSLOGD_CONF=${SYSLOGD_CONF:-"/etc/syslogd.conf"}

detect_syslog_daemon() {
  if [ -d "$RSYSLOG_DIR" ]; then
    syslog_daemon="rsyslog"
    config_dir=$RSYSLOG_DIR
  elif [ -d "$SYSLOGNG_DIR" ]; then
    syslog_daemon="syslog-ng"
    config_dir=$SYSLOGNG_DIR
  elif [ -f "$SYSLOGD_CONF" ]; then
    echo "Detected legacy syslogd, which only supports port 514."
    show_troubleshooting
    exit 2
  else
    echo "Could not detect syslog daemon in use."
    show_troubleshooting
    exit 2
  fi

  config_file=$config_dir/99-solarwinds.conf

  echo "    Detected syslog daemon: ${syslog_daemon}"
  echo "    Config file to be updated: ${config_file}"
}

detect_certificates() {
  if [ "$CA_PATH" ]; then
    echo "    Using CA_PATH override: ${CA_PATH}"
    return
  elif [ "$syslog_daemon" = rsyslog ]; then
    for CA_PATH in \
      /etc/ssl/certs/ca-certificates.crt \
      /etc/pki/tls/certs/ca-bundle.crt \
      /etc/ssl/ca-bundle.pem
    do
      if [ -e "$CA_PATH" ]; then
        echo "    Detected CA file: ${CA_PATH}"
        return
      fi
    done
  elif [ "$syslog_daemon" = syslog-ng ]; then
    CA_PATH=/etc/ssl/certs
    for crt in "$CA_PATH"/*.0; do
      if [ -e "$crt" ]; then
        echo "    Detected CA directory: ${CA_PATH}"
        return
      fi
    done
  fi

  echo "Could not auto-detect CA path. Consider installing ca-certificates,"
  echo "ca-certificates-bundle, or another similar package first."
  show_troubleshooting
  exit 2
}

detect_packages() {
  if [ "$syslog_daemon" = rsyslog ]; then
    dpkg-query -s rsyslog-openssl > /dev/null 2>&1
    if [ $? = 1 ]; then
      echo "    Will install DEB package: rsyslog-openssl"
      INSTALL_CMD="apt-get --trivial-only install rsyslog-openssl"
      return
    fi

    rpm -q rsyslog-openssl > /dev/null 2>&1
    if [ $? = 1 ]; then
      echo "    Will install RPM package: rsyslog-openssl"
      INSTALL_CMD="yum -yx rsyslog install rsyslog-openssl"
      return
    fi
  fi
}
INSTALL_CMD=

detect_writable_config() {
  if [ ! -w "$config_dir" ]; then
    echo "$config_dir is not writable. Please re-run this script as root or by using sudo."
    exit 2
  fi
}

detect_swo_config() {
  echo
  read_env_var SWO_HOSTNAME
  read_env_var SWO_PORT
  read_env_var SWO_TOKEN

  echo "    SolarWinds log endpoint: ${SWO_HOSTNAME}:${SWO_PORT}"
  echo "    SolarWinds token:"
  echo "        ${SWO_TOKEN}"
}

confirm_environment_detection() {
  echo "Auto-detected settings:"
  detect_syslog_daemon
  detect_certificates
  detect_packages
  detect_writable_config
  detect_swo_config
  echo

  if [ -z "$quiet_mode" ]; then
    # shellcheck disable=SC3045
    # Not POSIX-compliant, `read -p` works in both bash and Ubuntu's dash
    read -r -p "Is this correct? (y/n) " answer < /dev/tty
    echo
    case $answer in
      [Yy]*) return ;;
      [Nn]*) exit ;;
      *) echo "Please answer y or n." ;;
    esac
  fi
}

read_env_var() {
  if eval [ -z "\$$1" ]; then
    echo "What is the $1?"
    read -r "$1" < /dev/tty
    if eval [ -z "\$$1" ]; then
      exit
    fi
    echo
  fi
}

save_backup_config() {
  if [ -f "$config_file" ]; then
    echo
    echo "Detected an existing SolarWinds config. Backing it up to ${config_file}.bak"
    echo
    mv "$config_file" "$config_file.bak"
  fi
}

tls_rsyslog_config() {
  echo "\$DefaultNetstreamDriverCAFile ${CA_PATH}"
  echo "\$ActionSendStreamDriver ossl"
  echo "\$ActionSendStreamDriverMode 1"
  echo "\$ActionSendStreamDriverAuthMode x509/name"
  echo "\$ActionSendStreamDriverPermittedPeer *.${SWO_HOSTNAME#*.}"
  echo
  echo "\$template SWOFormat,\"<%PRI%>1 %TIMESTAMP:::date-rfc3339% %HOSTNAME% %APP-NAME% %PROCID% %MSGID% [${SWO_TOKEN}@41058]%msg:::sp-if-no-1st-sp%%msg%\""
  echo
  echo "*.*          @@${SWO_HOSTNAME}:${SWO_PORT};SWOFormat"
}

tls_syslogng_config() {
  # Using printf avoids echo's implementation-defined behavior for backslashes.
  printf "%s\n" 'template SWOFormat "<$PRI>1 $ISODATE ${HOST:--} ${PROGRAM:--} ${PID:--} ${MSGID:--}'" [${SWO_TOKEN}@41058] "'$MSG\n";'
  echo
  echo "destination d_solarwinds {"
  echo "  tcp(\"${SWO_HOSTNAME}\" port(${SWO_PORT})"
  echo "  template(SWOFormat)"
  echo "  tls(ca_dir(\"${CA_PATH}\") sni(yes)) );"
  echo "};"
  echo
  echo "log { source(s_src); destination(d_solarwinds); };"
}

write_rsyslog_config() {
  save_backup_config
  tls_rsyslog_config > "$config_file"
}

write_syslogng_config() {
  save_backup_config
  tls_syslogng_config > "$config_file"
}

restart_daemon() {
  svc=$(command -v service)
  if [ ! -f "$svc" ]; then
    echo
    echo "Cannot find service binary to restart ${syslog_daemon}. Restart service manually."
    echo
  else
    service "$syslog_daemon" restart
  fi
}

show_troubleshooting() {
  cat << EOF

This seems like a job for a human! Please either:
  • Let us make quick work of it. Copy and paste this output into an email
    to SWO-support@solarwinds.com.
  or
  • Perform the steps from the Add Data wizard manually.

EOF
}

show_help() {
  cat << EOF

Usage: ${0##*/} [-hvq]
  Setup system syslog daemon to log to SolarWinds.
  Supported syslog daemons: rsyslog, syslog-ng

  More: https://documentation.solarwinds.com/en/success_center/observability/

  Optional arguments:
      -h          display this help and exit
      -q          quiet/unattended mode; do not prompt for confirmation
      -V          display version and exit
EOF
}

main() {
  confirm_environment_detection

  if [ "$INSTALL_CMD" ]; then
    echo "Installing prerequisites: ${INSTALL_CMD}"
    if ! $INSTALL_CMD; then
      show_troubleshooting
      exit 2
    fi
    echo
  fi

  echo "Applying config and restarting ${syslog_daemon}..."
  if [ "$syslog_daemon" = rsyslog ]; then
    write_rsyslog_config
    restart_daemon
  elif [ "$syslog_daemon" = syslog-ng ]; then
    write_syslogng_config
    restart_daemon
  fi
}

while getopts ":hVq" opt; do
  case "$opt" in
  h)
    show_help >&2
    exit 1
    ;;
  q)
    quiet_mode=1
    ;;
  V)
    echo "$VERSION"
    exit 1
    ;;
  esac
done

main

echo
echo "SolarWinds setup complete!"
logger "SolarWinds setup complete for $(hostname)"

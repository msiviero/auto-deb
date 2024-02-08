#!/bin/bash

version=$GH_TAG_NAME
program_name=eliza-grpc-server
architecture=amd64
build_directory=./build-linux/$program_name-$version

echo "Preparing build directory"
rm -rf build-linux
mkdir -p $build_directory/DEBIAN
mkdir -p $build_directory/usr/local/bin
mkdir -p $build_directory/etc/systemd/system

echo "Compiling application executable" 
dart pub get
dart compile exe --define environment=production,version=$version bin/main.dart -o $build_directory/usr/local/bin/${program_name}

cat <<EOF >> $build_directory/DEBIAN/control
Package: $program_name
Architecture: all
Maintainer: Marco Siviero <m.siviero83@gmail.com>
Priority: optional
Version: $version
Description: Eliza grpc server
EOF

cat <<EOF >> $build_directory/etc/systemd/system/app.service
[Unit]
Description=Application
After=network.target
[Service]
Environment="ELIZA_GRPC_SERVER_CERT=/etc/letsencrypt/live/app.eliza.cool/fullchain.pem"
Environment="ELIZA_GRPC_SERVER_PRIVATE_KEY=/etc/letsencrypt/live/app.eliza.cool/privkey.pem"
Environment="ELIZA_STATIC_DIR=/home/app/www"
ExecStart=/usr/local/bin/${program_name}
WorkingDirectory=/home/app
Restart=always
RestartSec=1
User=app
LimitNOFILE=640000
[Install]
WantedBy=multi-user.target
EOF

cat <<EOF >> $build_directory/DEBIAN/preinst
#!/bin/bash
if systemctl is-active --quiet app.service
then
  echo "daemon is running,stopping it"
  systemctl stop app.service
else
  echo "daemon is not running"
fi
EOF
chmod 755 $build_directory/DEBIAN/preinst

cat <<EOF >> $build_directory/DEBIAN/postinst
#!/bin/bash
systemctl enable app.service
systemctl start app.service
EOF
chmod 755 $build_directory/DEBIAN/postinst

cat <<EOF >> $build_directory/DEBIAN/prerm
#!/bin/bash
systemctl stop app.service
systemctl disable app.service
EOF
chmod 755 $build_directory/DEBIAN/prerm

dpkg-deb --build $build_directory
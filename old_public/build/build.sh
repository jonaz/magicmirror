rm ../dist/* -rf

#if [[ ! -f ../scripts/config.js ]]; then
#    echo ""
#    echo "==== [ Copy default config file ] =================="
#    cp -v ../scripts/config.example.js ../scripts/config.js
#fi

echo ""
echo "==== [ Compile APP ] =================="
r.js -o app.build.js


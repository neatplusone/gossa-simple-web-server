# Gossa usage POST example automation curl powershell:  
  
When server runs with argument or to recreate demo:
```
mkdir rwshare
./gossa -h 0.0.0.0 -p 8080 -prefix /rwhttp/ rwshare
```

Client - Powershell with CURL and Native Post example:
```ps1
# CURL Windows PS Example
$OutputFile = "testfile.txt"
$UploadAPI_Endpoint = "http://gossa.lan:8080"
$UploadAPI_Directory = "rwhttp"
 
curl.exe -v -X POST -H "gossa-path: /$UploadAPI_Directory/$OutputFile" -F "$OutputFile=@$OutputFile" $UploadAPI_Endpoint/$UploadAPI_Directory/post
```
```sh
# CURL Linux
export UploadAPI_Directory="rwhttp"
export OutputFile="testfile.txt"
export UploadAPI_Endpoint="http://gossa.lan:8080"

curl -v -X POST \
  -H "gossa-path: /${UploadAPI_Directory}/${OutputFile}" \
  -F "${OutputFile}=@${OutputFile}" \
  "${UploadAPI_Endpoint}/${UploadAPI_Directory}/post"
``` 

```cmd
set UploadAPI_Directory=rwhttp
set OutputFile=file.txt
set UploadAPI_Endpoint=http://gossa.lan:8080

curl.exe -v -X POST ^
  -H "gossa-path: /%UploadAPI_Directory%/%OutputFile%" ^
  -F "%OutputFile%@=%OutputFile%" ^
  %UploadAPI_Endpoint%/%UploadAPI_Directory%/post
```

```ps1
# needs to be redeveloped / tested, maybe even simplify it by changing how server handles POST entries.
# likely when / if WebDav gets added
```

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
REM Windows CMD / Batch file
set UploadAPI_Directory=rwhttp
set OutputFile=file.txt
set UploadAPI_Endpoint=http://gossa.lan:8080

curl.exe -v -X POST ^
  -H "gossa-path: /%UploadAPI_Directory%/%OutputFile%" ^
  -F "%OutputFile%@=%OutputFile%" ^
  %UploadAPI_Endpoint%/%UploadAPI_Directory%/post
```

```ps1
# CURL and Native Windows PS Example
$GossaSWSUploadThis = "testfile.txt"
$UploadAPI_Endpoint = "http://gossa.lan:8080/rwhttp"

write-host "SEND CURL"
curl.exe -v -X POST -H "gossa-path: $GossaSWSUploadThis" -F "$GossaSWSUploadThis=@$GossaSWSUploadThis" $UploadAPI_Endpoint/post

write-host "SEND GOSSA"
function Upload-Gossa ($filePath) {
    $file = Get-Item $filePath
    $fileName = $file.Name

    $uri = "$UploadAPI_Endpoint/post"
    
    $boundary = [System.Guid]::NewGuid().ToString()
    $LF = "`r`n"
    $preFile = (
        "--$boundary",
        "Content-Disposition: form-data; name=`"file`"; filename=`"$fileName`"",
        "Content-Type: application/octet-stream$LF$LF"
    ) -join $LF
    $postFile = "$LF--$boundary--$LF"

    $preBytes = [System.Text.Encoding]::UTF8.GetBytes($preFile)
    $fileBytes = [System.IO.File]::ReadAllBytes($file.FullName)
    $postBytes = [System.Text.Encoding]::UTF8.GetBytes($postFile)

    $bodyBytes = New-Object byte[] ($preBytes.Length + $fileBytes.Length + $postBytes.Length)
    [System.Buffer]::BlockCopy($preBytes, 0, $bodyBytes, 0, $preBytes.Length)
    [System.Buffer]::BlockCopy($fileBytes, 0, $bodyBytes, $preBytes.Length, $fileBytes.Length)
    [System.Buffer]::BlockCopy($postBytes, 0, $bodyBytes, $preBytes.Length + $fileBytes.Length, $postBytes.Length)

    $headers = @{
        "Content-Type" = "multipart/form-data; boundary=$boundary"
        "gossa-path" = "$fileName"  # adjust path as needed
    }

    Invoke-RestMethod -Uri $uri -Method Post -Headers $headers -Body $bodyBytes
}

Upload-Gossa $GossaSWSUploadThis
```

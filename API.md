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
# Native Powershell implementation function
$OutputFile = "testfile.txt"
$UploadAPI_Endpoint = "http://gossa.lan:8080"
$UploadAPI_Directory = "rwhttp"

function Send-GossaFile {
    [CmdletBinding()]
    param(
        # Local file path to upload
        [Parameter(Mandatory=$true)]
        [string]$FilePath,
 
        # Full base URL up to the /rwhttp path
        [Parameter(Mandatory=$true)]
        [string]$BaseUrl = "$UploadAPI_Endpoint/$UploadAPI_Directory"
    )
 
    # Extract just the file name from the path
    $FileName = [System.IO.Path]::GetFileName($FilePath)
 
    # "gossa-path" must point to the folder or subpath on the server
    # Because JS sets "gossa-path: /$UploadAPI_Directory/$OutputFile"
    # we do the same here:
    $headers = @{
        "gossa-path" = "/$UploadAPI_Directory/$FileName"
    }
 
    # For the form fields, the key (the "name" in form-data) must match
    # $FileName, just like browser curl output had "testfile.txt=@testfile.txt" in cURL
    $formData = @{
        $FileName = Get-Item -Path $FilePath
    }
 
    # The final POST endpoint is basically $BaseUrl + "/post"
    $postUrl = Join-Path -Path $BaseUrl -ChildPath "post"
 
    Write-Verbose "Posting to $postUrl with file name = $FileName"
 
    # Send multipart/form-data
    Invoke-WebRequest `
        -Uri $postUrl `
        -Method Post `
        -Headers $headers `
        -Form $formData `
        -Verbose
}
 
Send-GossaFile -FilePath $OutputFile -BaseUrl "$UploadAPI_Endpoint/$UploadAPI_Directory"
```
# Script that downloads a file from the server to the output location ignoring the server certificate
param (
    [Parameter(Mandatory=$true)][string]$server,
    [Parameter(Mandatory=$true)][string]$output,
    [Parameter(Mandatory=$true)][string]$useragent
)

# Prevent the progress meter from accessing the console.
$ProgressPreference = "SilentlyContinue"

if (-not("dummy" -as [type])) {
    add-type -TypeDefinition @"
using System;
using System.Net;
using System.Net.Security;
using System.Security.Cryptography.X509Certificates;
public static class Dummy {
    public static bool ReturnTrue(object sender,
        X509Certificate certificate,
        X509Chain chain,
        SslPolicyErrors sslPolicyErrors) { return true; }
    public static RemoteCertificateValidationCallback GetDelegate() {
        return new RemoteCertificateValidationCallback(Dummy.ReturnTrue);
    }
}
"@
}
[System.Net.ServicePointManager]::ServerCertificateValidationCallback = [dummy]::GetDelegate()

# $null is needed to prevent wget from attempting read the standard input or
# output streams when attached to the console.
$null | wget -UserAgent $useragent $server -o $output > $null

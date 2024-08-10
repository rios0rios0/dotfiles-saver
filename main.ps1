param (
    [string]$operation
)

# Get the current Windows user folder
$currentUser = [System.Environment]::UserName
$backupPath = "C:\Users\$currentUser\OneDrive\Backup"

# Detect WSL distros and get the default one
$defaultWSL = wsl -l -v | Select-String -Pattern "\*" | ForEach-Object { $_.Line.Split(" ")[1] }

# Define the source and destination map with specific files
$paths = @{
    "win" = @(
        # folders group
        ".aws\config",
        ".aws\credentials",
        ".azure\azureProfile.json",
        ".azure\service_principal_entries.json",
        ".gnupg",
        ".ssh",
        #"Development",

        # files group
        ".gitconfig",
        ".gitignore",
        ".oh-my-posh.json",
        ".wakatime.cfg"
        #"AppData\Local\Packages\Microsoft.WindowsTerminal_*\LocalState\settings.json"
    )
    "wsl" = @(
        # folders group
        ".docker\config.json",
        ".john",
        ".kube\config",
        ".kube\config-files",
        ".sqlmap",
        ".zsh_history_list",
        #"Development",

        # files group
        ".autobump.yaml",
        ".bashrc",
        ".freterc",
        ".gitconfig",
        ".gitignore",
        ".npmrc",
        ".npmrc.vizir",
        ".p10k.zsh",
        ".zsh_history",
        ".zshrc",
        "pyvenv.cfg"
    )
}

function CopyFiles {
    param (
        [string]$source,
        [string]$destination,
        [string[]]$items,
        [string]$operation
    )

    function CopyItemRecursively {
        param (
            [string]$sourcePath,
            [string]$destinationPath,
            [string]$operation
        )

        # Remove null characters from the source path (Illegal characters in path)
        $cleanSourcePath = $sourcePath -replace "`0", ""
        if (Test-Path -Path $cleanSourcePath.Trim() -PathType Container) {
            # Ensure the destination folder exists
            if (-not (Test-Path $destinationPath)) {
                New-Item -Path $destinationPath -ItemType Directory -Force
            }
            Get-ChildItem -Path $cleanSourcePath | ForEach-Object {
                CopyItemRecursively -sourcePath $_.FullName -destinationPath "$destinationPath\$($_.Name)" -operation $operation
            }
            Write-Output "$operation folder $sourcePath to $destinationPath"
        } else {
            # Ensure the parent folder exists in the destination
            $parentFolder = Split-Path -Path $destinationPath -Parent
            if (-not (Test-Path $parentFolder)) {
                New-Item -Path $parentFolder -ItemType Directory -Force
            }
            Copy-Item -Path $cleanSourcePath.Trim() -Destination "$destinationPath" -Force
            Write-Output "$operation file $sourcePath to $destinationPath"
        }
    }

    foreach ($item in $items) {
        $sourcePath = "$source\$item"
        $destinationPath = "$destination\$item"
        CopyItemRecursively -sourcePath $sourcePath -destinationPath $destinationPath -operation $operation
    }
}

# Perform the requested operation
switch ($operation) {
    "backup" {
        # Backup WIN files from root user folder in Windows
        CopyFiles -source "C:\Users\$currentUser" -destination $backupPath\win -items $paths["win"] -operation "Backed up"

        # Backup WSL files from default WSL path
        CopyFiles -source "\\wsl.localhost\$defaultWSL\home\$currentUser" -destination $backupPath\wsl -items $paths["wsl"] -operation "Backed up"
    }
    "restore" {
        # Restore WIN files to root user folder in Windows
        CopyFiles -source $backupPath\win -destination "C:\Users\$currentUser" -items $paths["win"] -operation "Restored"

        # Restore WSL files to default WSL path
        CopyFiles -source $backupPath\wsl -destination "\\wsl.localhost\$defaultWSL\home\$currentUser" -items $paths["wsl"] -operation "Restored"
    }
    default {
        Write-Output "Invalid operation. Use 'backup' or 'restore'."
    }
}

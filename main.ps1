param (
    [string]$operation
)

# Get the current Windows user folder
$user = [System.Environment]::UserName
$backupPath = "C:\Users\$user\OneDrive\Backup"

# Detect WSL distros and get the default one
$defaultWSL = wsl -l -v | Select-String -Pattern "\*" | ForEach-Object { $_.Line.Split(" ")[1] }

# Define the source and destination map with specific files
$paths = @{
    "win" = @(
        # folders or recursive group
        ".aws\config",
        ".aws\credentials",
        ".azure\azureProfile.json",
        ".azure\service_principal_entries.json",
        ".gnupg",
        ".ssh",
        "AppData\Local\Packages\Microsoft.WindowsTerminal_*\LocalState\settings.json",
        #"Development",

        # direct files group
        ".gitconfig",
        ".gitignore",
        ".oh-my-posh.json",
        ".wakatime.cfg"
    )
    "wsl" = @(
        # folders or recursive group
        ".docker\config.json",
        ".histdb",
        ".john",
        ".kube\config",
        ".kube\config-files",
        ".sqlmap",
        #"Development",

        # direct files group
        ".autobump.yaml",
        ".freterc",
        ".gitconfig",
        ".gitignore",
        ".npmrc",
        ".npmrc.vizir",
        ".p10k.zsh",
        ".zshrc",
        "pyvenv.cfg" # TODO: do I really need to backup this file?
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
        $resolvedItem = ResolveWildcardPath -path "$source\$item"
        $relativePath = $resolvedItem.Substring($source.Length)
        $sourcePath = "$source\$relativePath"
        $destinationPath = "$destination\$relativePath"
        CopyItemRecursively -sourcePath $sourcePath -destinationPath $destinationPath -operation $operation
    }
}

function ResolveWildcardPath {
    param (
        [string]$path
    )

    if ($path -like '*[*]*') {
        $resolvedPaths = Resolve-Path -Path $path
        if ($resolvedPaths) {
            return $resolvedPaths[0].Path
        }
    }

    return $path
}

# Perform the requested operation
switch ($operation) {
    "backup" {
        # Backup WIN files from root user folder in Windows
        CopyFiles -source "C:\Users\$user" -destination $backupPath\win -items $paths["win"] -operation "Backed up"

        # Backup WSL files from default WSL path
        CopyFiles -source "\\wsl.localhost\$defaultWSL\home\$user" -destination $backupPath\wsl -items $paths["wsl"] -operation "Backed up"
    }
    "restore" {
        # Restore WIN files to root user folder in Windows
        CopyFiles -source $backupPath\win -destination "C:\Users\$user" -items $paths["win"] -operation "Restored"

        # Restore WSL files to default WSL path
        CopyFiles -source $backupPath\wsl -destination "\\wsl.localhost\$defaultWSL\home\$user" -items $paths["wsl"] -operation "Restored"
    }
    default {
        Write-Output "Invalid operation. Use 'backup' or 'restore'."
    }
}

# TODO: change this script to avoid copying "node_modules", ".venv", ".terraform", ".terragrunt-cache" folders
# download the dotfiles from the github.com/user/dotfiles repository, where user is the current user
# inject 1Password credentials in the ones that are not public

#Install-Module -Name Pester -Scope CurrentUser -Force -SkipPublisherCheck -DestinationPath .\Modules

# Update the PSModulePath to include the local module directory
#$env:PSModulePath = "$PSScriptRoot\Modules;" + $env:PSModulePath

# Import Pester module
Import-Module Pester

# Load the script to be tested
. "$PSScriptRoot\main.ps1"

Describe "CopyFiles Function" {
    It "should copy files from source to destination" {
        $source = "C:\TestSource"
        $destination = "C:\TestDestination"
        $items = @("file1.txt", "file2.txt")
        $operation = "Backed up"

        # Create test files
        New-Item -Path "$source\file1.txt" -ItemType File -Force
        New-Item -Path "$source\file2.txt" -ItemType File -Force

        # Run the function
        CopyFiles -source $source -destination $destination -items $items -operation $operation

        # Assert the files were copied
        Test-Path "$destination\file1.txt" | Should -Be $true
        Test-Path "$destination\file2.txt" | Should -Be $true
    }

    It "should copy folders from source to destination" {
        $source = "C:\TestSource"
        $destination = "C:\TestDestination"
        $items = @("folder1", "folder2")
        $operation = "Backed up"

        # Create test folders
        New-Item -Path "$source\folder1" -ItemType Directory -Force
        New-Item -Path "$source\folder2" -ItemType Directory -Force

        # Run the function
        CopyFiles -source $source -destination $destination -items $items -operation $operation

        # Assert the folders were copied
        Test-Path "$destination\folder1" | Should -Be $true
        Test-Path "$destination\folder2" | Should -Be $true
    }

    It "should create destination folder if it does not exist" {
        $source = "C:\TestSource"
        $destination = "C:\TestDestination"
        $items = @("folder1")
        $operation = "Backed up"

        # Create test folder
        New-Item -Path "$source\folder1" -ItemType Directory -Force

        # Run the function
        CopyFiles -source $source -destination $destination -items $items -operation $operation

        # Assert the destination folder was created
        Test-Path "$destination\folder1" | Should -Be $true
    }

    It "should overwrite existing files in the destination" {
        $source = "C:\TestSource"
        $destination = "C:\TestDestination"
        $items = @("file1.txt")
        $operation = "Backed up"

        # Create test files
        New-Item -Path "$source\file1.txt" -ItemType File -Force
        New-Item -Path "$destination\file1.txt" -ItemType File -Force

        # Write different content to the source file
        Set-Content -Path "$source\file1.txt" -Value "New content"

        # Run the function
        CopyFiles -source $source -destination $destination -items $items -operation $operation

        # Assert the file was overwritten
        (Get-Content -Path "$destination\file1.txt") | Should -Be "New content"
    }
}

Describe "Switch Operation" {
    It "should perform backup operation" {
        $operation = "backup"

        # Mock the CopyFiles function
        Mock -CommandName CopyFiles -MockWith { }

        # Run the script
        . "$PSScriptRoot\main.ps1" -operation $operation

        # Assert the CopyFiles function was called with the correct parameters
        Assert-MockCalled -CommandName CopyFiles -Exactly 2 -Scope It
    }

    It "should perform restore operation" {
        $operation = "restore"

        # Mock the CopyFiles function
        Mock -CommandName CopyFiles -MockWith { }

        # Run the script
        . "$PSScriptRoot\backup_restore.ps1" -operation $operation

        # Assert the CopyFiles function was called with the correct parameters
        Assert-MockCalled -CommandName CopyFiles -Exactly 2 -Scope It
    }
}

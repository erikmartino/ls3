# Homebrew-ls3 Repository Setup Instructions

This guide will help you set up the `homebrew-ls3` repository to enable Homebrew installation of your `ls3` application.

## Step 1: Create the GitHub Repository

1. Go to https://github.com/new
2. Set the repository name to: `homebrew-ls3`
3. Set it to **Public** (required for Homebrew taps)
4. **Do NOT** initialize with README, .gitignore, or license (we'll add these manually)
5. Click "Create repository"

## Step 2: Initialize the Repository Locally

```bash
# Create a new directory for the homebrew tap
mkdir homebrew-ls3
cd homebrew-ls3

# Initialize git repository
git init
git branch -M main

# Copy the files from this setup directory
cp ../homebrew-ls3-setup/README.md .
cp -r ../homebrew-ls3-setup/Formula .

# Add and commit the files
git add .
git commit -m "Initial commit: Add README and Formula directory"

# Add the remote origin (replace 'erikmartino' with your GitHub username if different)
git remote add origin https://github.com/erikmartino/homebrew-ls3.git

# Push to GitHub
git push -u origin main
```

## Step 3: Set Up GitHub Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Fine-grained tokens
2. Click "Generate new token"
3. Configure the token:
   - **Name**: `GoReleaser Homebrew Token`
   - **Expiration**: Choose your preferred expiration (1 year recommended)
   - **Repository access**: Select repositories → Choose `ls3` and `homebrew-ls3`
   - **Repository permissions**:
     - Contents: Read and Write
     - Metadata: Read
     - Pull requests: Read and Write

4. Generate the token and copy it immediately

## Step 4: Configure the Token Locally

Set the token as an environment variable in your shell profile:

```bash
# Add to your ~/.bashrc, ~/.zshrc, or ~/.bash_profile
export GITHUB_TOKEN="your_token_here"

# Or set it temporarily for the current session
export GITHUB_TOKEN="your_token_here"
```

## Step 5: Test the Setup

1. Go back to your main `ls3` repository:
   ```bash
   cd ../ls3
   ```

2. Create a test release:
   ```bash
   # Create a new tag (increment from your latest version)
   git tag v1.3.1
   git push origin v1.3.1
   
   # Run GoReleaser to create the release
   goreleaser release --clean
   ```

3. Check that:
   - A new release was created on GitHub
   - The homebrew formula was automatically added to your `homebrew-ls3` repository
   - The formula file is located at `Formula/ls3.rb`

## Step 6: Verify Installation

Test that users can install your application:

```bash
# Add your tap
brew tap erikmartino/ls3

# Install ls3
brew install ls3

# Test that it works
ls3 --help
```

## Troubleshooting

### Token Issues
- Make sure the token has the correct permissions
- Verify the token is set as `GITHUB_TOKEN` environment variable
- Check that both repositories (`ls3` and `homebrew-ls3`) are included in the token scope

### Repository Issues
- Ensure the `homebrew-ls3` repository is public
- Verify the repository name follows the `homebrew-<name>` convention
- Make sure the Formula directory exists (even if empty initially)

### Release Issues
- Check that your git working directory is clean
- Verify that the tag was pushed to the remote repository
- Review GoReleaser logs for any error messages

## Post-Setup

Once everything is working:

1. Update your main `ls3` repository README to include Homebrew installation instructions
2. Consider adding the Homebrew installation method to your release notes
3. Test the installation process on a clean system to ensure it works for end users

## Future Releases

For future releases, simply:

1. Create and push a new tag: `git tag v1.x.x && git push origin v1.x.x`
2. Run GoReleaser: `goreleaser release --clean`
3. GoReleaser will automatically update the Homebrew formula

That's it! Your application will be available via Homebrew.
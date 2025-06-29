# Railway Deployment Guide

## Quick Deploy to Railway

### 1. Fork/Clone Repository
- Fork this repository to your GitHub account
- Or clone it to your local machine

### 2. Connect to Railway
1. Go to [Railway.app](https://railway.app)
2. Click "New Project"
3. Select "Deploy from GitHub repo"
4. Choose your forked repository

### 3. Configure Environment Variables
In Railway dashboard, go to your project → Variables tab and add:

```
GEMINI_API_KEY=your_actual_gemini_api_key_here
```

### 4. Deploy
Railway will automatically:
- Detect the Dockerfile
- Build the Docker image
- Deploy to port 8008
- Provide a public URL

### 5. Test Your Deployment
```bash
# Test health endpoint
curl https://your-app-name.up.railway.app/health

# Test the API
curl -X POST https://your-app-name.up.railway.app/prompt \
  -H "Content-Type: application/json" \
  -d '{"session_id": "test123", "prompt": "Hello! What can you help me with?"}'
```

## Railway Configuration

### Port Configuration
- Railway automatically assigns a `PORT` environment variable
- The Dockerfile exposes port 8008
- Railway will map the external port to your app

### Environment Variables
- `GEMINI_API_KEY`: Required - Your Gemini API key
- `PORT`: Automatically set by Railway (optional override)

### Health Checks
Railway will automatically check the `/health` endpoint to ensure your service is running.

## Troubleshooting

### "Application not found" Error
- Ensure your Railway project is properly deployed
- Check that the service is running in Railway dashboard
- Verify the URL is correct

### API Key Issues
- Ensure `GEMINI_API_KEY` is set in Railway environment variables
- Check that the API key is valid and has proper permissions

### Build Issues
- Check Railway build logs for any compilation errors
- Ensure all required files (main.go, config.go) are in the repository 
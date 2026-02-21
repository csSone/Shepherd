# Shepherd Model Management - Verification Plan

## Goal
Verify that the model management fixes (Phases 1-4) are working correctly by checking logs, testing API endpoints, and confirming models appear in the web UI.

---

## Task 1: Check Latest Log File
- [ ] **Find and read the most recent log file** in `/home/user/workspace/Shepherd/logs/`
- [ ] **Look for**:
  - Model manager initialization messages
  - PathConfigs loading
  - Any errors during model scanning
  - API server startup
- **Verification**: Can see log entries showing model paths loaded

---

## Task 2: Test API Endpoints
- [ ] **Test GET /api/models endpoint** using curl
  - Expected: JSON response with ModelDTO array
  - Check for proper serialization (no empty fields)
- [ ] **Test POST /api/models/scan endpoint** 
  - Expected: Trigger scan, return success
- [ ] **Verify response structure** matches ModelDTO format
- **Verification**: API returns valid JSON with model data

---

## Task 3: Verify Configuration
- [ ] **Read config file** `/home/user/workspace/Shepherd/config/server.config.yaml`
- [ ] **Verify path_configs** contains `/home/user/workspace/LlamacppServer/build/models`
- [ ] **Check model directory** exists and has subdirectories with .gguf files
- **Verification**: Config and filesystem state are correct

---

## Task 4: Check Build and Start Server (if needed)
- [ ] **Build the backend** with `make build`
- [ ] **Start server** with `./build/shepherd standalone`
- [ ] **Verify server starts** without errors
- **Verification**: Server running and accepting requests

---

## Task 5: Document Findings
- [ ] **Summarize log analysis**
- [ ] **Document API test results**
- [ ] **Identify any remaining issues** that need fixing
- [ ] **Recommend next steps** (if fixes needed)
- **Verification**: Clear report of current state

---

## Configuration Details

**Config path**: `/home/user/workspace/Shepherd/config/server.config.yaml`
**Log directory**: `/home/user/workspace/Shepherd/logs/`
**Model directory**: `/home/user/workspace/LlamacppServer/build/models/`
**API Base URL**: `http://localhost:8080`

**Model subdirectories to check**:
- DavidAU
- noctrex
- Qwen
- seanbailey518
- unsloth

---

## Success Criteria
1. Logs show model paths being loaded correctly
2. `/api/models` returns proper JSON with model data
3. No errors in log files related to model scanning
4. Configuration is correctly set up

## Potential Issues to Watch For
- Empty model list in API response
- Serialization errors (null/empty fields)
- Path resolution issues
- File permission errors
- Missing .gguf files in subdirectories

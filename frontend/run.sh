# !/bin/bash
npm audit fix --force
npm install . --force
npm run dev
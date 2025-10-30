# Start the backend
cd backend
./connect-four &
BACKEND_PID=$!

# Wait for backend to start
sleep 5

# Start the frontend
cd ../frontend

# Wait for both processes
wait $BACKEND_PID
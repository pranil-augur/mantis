# Use an official Python runtime as a parent image
FROM python:3.10-slim

# Set the working directory in the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . .

# Install any needed packages specified in requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Make port 80 available to the world outside this container
EXPOSE 80

# Define environment variables for database connection
ENV DB_HOST=
ENV DB_NAME=
ENV DB_USER=
ENV DB_PASSWORD=

# Run app.py when the container launches
CMD ["python", "app.py"]


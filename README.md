# SiteReliabilityEngineering-glints-Task

Cold Storage Group Ordering
Cold Storage is a popular supermarket chain in Singapore. As part of an employee benefit, Glints employees located in Singapore are given a fixed weekly office budget to purchase groceries to fill the office pantry. Unfortunately, the Cold Storage website does not have a group ordering function, which makes collating the list of items to buy a hassle.

To solve this problem, you proposed creating a service that everyone in the office can connect to, that displays the Cold Storage website in a headless browser. This simulates group ordering functionality. This service contains the following components:

A headless browser, such as Google Chrome
A VNC server, such as TigerVNC that allows for remote desktop functionality
noVNC, a JavaScript-based VNC client that allows easy access through a URL
The Task
The task is to containerize and deploy the above components onto a Kubernetes cluster. You should ensure that the deployment is as declarative as possible, ensuring that all files required to re-create the deployment are included.

For the purpose of this assessment, there is no need to use a publicly hosted Kubernetes cluster; you may use Minikube to quickly spin up a local Kubernetes cluster. If you require a Docker Registry to host Docker images, you may use the public Docker Hub.

Roughly speaking, these are the sub-tasks:

Write one or more Dockerfiles for the 3 services above.
Build the appropriate images and push them to a Docker Registry.
Write one or more Kubernetes YAML files to deploy the images to the Kubernetes cluster.
At the end, employees should be able to access a URL which will display the contents of the headless browser.

Submission
Include a repository that includes the Dockerfiles have you written, as well as any relevant Kubernetes YAML files. Take care not to include any secrets.

Bonus
This is optional, and serves as additional proof points. We will consider it complete even without this functionality.

Determine metrics that would reflect the reliability of this service.

Implement monitoring and alerting in case the service stops working. You may use any means of monitoring that you can think of.

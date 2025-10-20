Hi, and welcome to my Test 1 - The national in-service api Created by and developed by Victor Tillett

Project description:

This Project is a National Inservice Api ONlY project Given to us by  Dalwin Lewis on the 01/10/2025.

This project is to given to use either to be done in pairs of 2 or done Individually(My route) 

This project is to be completed in 3 checkpoints

Checkpoint 1: ERD Diagram Due October 6th 2025
Checkpoint 2 : API Part 1: CRUD (Create,Read,Update,Delete) With additional services Rate limit, CORS, pagination and Sorting AND Graceful shutdown Due October 15th 2025 - Status: Ontime 

Checkpoint 3/ Test 1 Submission - Final Github submission and Presentation 

Requirements (Go version, Docker)


  Quickstart Guide

PLease use the Latest version of docker to run this file 

locate the docker compose file in  test1-national-inservice-api

Then 
 type "docker compose up"

 Wait a while for the services to build after you are ready for testing and grading 

 now open another terminal 
  cd test1-national-inservice-api/ 

now 

cd app 

You are now in the root of the project to start testing the project in its entirety

To make the html file run on the Docker server please DISABLE THE DEFAULT APACHE SERVICE ON LINUX 

use this line to stop it if you want to view the file or else the the website will be blank 

sudo systemctl disable apache2

then renable it using
 
sudo systemctl enable apache2
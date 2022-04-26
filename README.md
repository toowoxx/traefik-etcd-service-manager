# Traefik Service Manager for etcd

This program runs indefinitely and tracks container changes from Docker.
These changes are then used to propagate Traefik labels into etcd which is then
used by a Traefik instance.

The use case here is for distributed systems where Traefik and etcd runs on a different machine
than the Docker containers that need the routes configured.

## License

All files in this repository are licensed under the Apache License 2.0 unless specified otherwise.

```
   Copyright 2022 Toowoxx IT GmbH

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
```

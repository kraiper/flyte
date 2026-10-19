from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ActionAddress(_message.Message):
    __slots__ = ["url", "pod", "cluster"]
    URL_FIELD_NUMBER: _ClassVar[int]
    POD_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    url: str
    pod: PodAddress
    cluster: str
    def __init__(self, url: _Optional[str] = ..., pod: _Optional[_Union[PodAddress, _Mapping]] = ..., cluster: _Optional[str] = ...) -> None: ...

class PodAddress(_message.Message):
    __slots__ = ["namespace", "name", "ip", "port"]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    IP_FIELD_NUMBER: _ClassVar[int]
    PORT_FIELD_NUMBER: _ClassVar[int]
    namespace: str
    name: str
    ip: str
    port: int
    def __init__(self, namespace: _Optional[str] = ..., name: _Optional[str] = ..., ip: _Optional[str] = ..., port: _Optional[int] = ...) -> None: ...

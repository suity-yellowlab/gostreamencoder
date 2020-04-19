Package suity-yellowlab/gostreamencoder provides a mechanism to encode the data of a stream over a continous stream connection that is not closed on finishing
without loading all the data into memory like gob or json encoding.
Use cases are e.g. sending m files over webrtc datachannels or keep alive tcp connections. Data integrity is not checked as this should be abstracted from the chunking.
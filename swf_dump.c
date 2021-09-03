/**
 * A simple program to dump information from uncompressed Adobe Shockwave
 * Flash (SWF) files. This program only supports uncompressed SWF files;
 * compressed ones (that have a signature of "CWS") will have to be
 * decompressed using zlib or LZMA first.
 *
 * Licensed under the Apache License.
 * (C) 2014 Wong Yong Jie.
 */

#include <stdio.h>
#include <stdint.h>
#include <math.h>

char* tag_types[1024];
void init_tag_types();

int main(int argc, char* argv[]) {
  init_tag_types();

  /* parse command line arguments */
  if (argc != 2) {
    printf("usage: ./swf_dump <file>\n");
    return 1;
  }

  /* open the file */
  FILE* file = fopen(argv[1], "rb");

  /* verify this is indeed an SWF file */
  char signature[4] = { '\0' };
  fread(&signature, 1, 3, file);
  if (strcmp(signature, "FWS") != 0) {
    printf("input is not an uncompressed SWF file\n");
    return 2;
  }

  /* skip past the fixed part of the header */
  fseek(file, 8, SEEK_SET);

  /* parse and skip the variable part of the header */
  uint8_t nbits;
  fread(&nbits, 1, 1, file);
  nbits = (nbits & 0xf8) >> 3;
  uint32_t frame_size_length_bits = 5 + (nbits * 4);
  uint32_t frame_size_length = ceil(frame_size_length_bits / 8.0);
  //printf("frame_size_length = %d\n", frame_size_length);
  fseek(file, frame_size_length - 1, SEEK_CUR);
  
  /* skip the frame rate and frame count */
  fseek(file, 4, SEEK_CUR);

  printf("%10s %8s %20s %10s\n", "offset", "tag_code", "name", "length");
  do {
    off_t offset = ftell(file);
    uint16_t tag_code_and_length;
    fread(&tag_code_and_length, 2, 1, file);

    /* tag code in upper 10 bits, length in lower 6 bits */
    uint16_t tag_code = (tag_code_and_length & 0xffc0) >> 6;
    uint32_t length = tag_code_and_length & 0x3f;
    
    /* tag code is 0 means eof */
    if (tag_code == 0) {
      break;
    }

    /* length is stored in the subsequent 4 bytes if length is 0x3f */
    if (length == 0x3f) {
      fread(&length, 4, 1, file);
    }

    printf("%#010lx %#04x     %20s %#010x\n", offset, tag_code,
        tag_types[tag_code], length);
    fseek(file, length, SEEK_CUR);

  } while (!feof(file));

  return 0;
}

void init_tag_types() {
  int i;

  /* initialize everything to Unknown */
  for (i = 0; i < 1024; i++) {
    tag_types[i] = "Unknown";
  }

  /* define the known types */
  tag_types[1] = "End";
  tag_types[2] = "DefineShape";
  tag_types[9] = "SetBackgroundColor";
  tag_types[20] = "DefineBitsLossless";
  tag_types[22] = "DefineShape2";
  tag_types[32] = "DefineShape3";
  tag_types[33] = "DefineText2";
  tag_types[34] = "DefineButton2";
  tag_types[36] = "DefineBitsLossless2";
  tag_types[37] = "DefineEditText";
  tag_types[39] = "DefineSprite";
  tag_types[43] = "FrameLabel";
  tag_types[56] = "ExportAssets";
  tag_types[65] = "ScriptLimits";
  tag_types[69] = "FileAttributes";
  tag_types[73] = "DefineFontAlignZones";
  tag_types[74] = "CSMTextSettings";
  tag_types[75] = "DefineFont3";
  tag_types[76] = "SymbolClass";
  tag_types[77] = "Metadata";
  tag_types[78] = "DefineScalingGrid";
  tag_types[82] = "DoABC";
  tag_types[83] = "DefineShape4";
  tag_types[88] = "DefineFontName";
}

/* vim: set ts=2 sw=2 et: */

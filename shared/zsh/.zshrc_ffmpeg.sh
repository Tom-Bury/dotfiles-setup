#!/usr/bin/env zsh

function 2gif() {
    input_file=$1

    if [ -z "$input_file" ]; then
        input_file=$(ls | fzf)
    fi

    while true; do
        echo "Enter scale value (e.g., 1.5 for 1.5 times the original size. Default 1):"
        read scale_value
        if [[ -z "$scale_value" ]]; then
            scale_value=1
            break
        elif [[ $scale_value =~ ^[0-9]+([.][0-9]+)?$ ]]; then
            break
        fi
    done

    while true; do
        echo "Limit colors? (y/n):"
        read limit_colors
        if [[ -z "$limit_colors" ]]; then
            limit_colors="y"
            break
        elif [[ $limit_colors == "y" || $limit_colors == "n" ]]; then
            break
        fi
    done

    while true; do
        echo "Enter frame rate (e.g. 16):"
        read frame_rate
        if [[ -z "$frame_rate" ]]; then
            frame_rate=16
            break
        elif [[ $frame_rate =~ ^[0-9]+$ ]]; then
            break
        fi
    done

    if [ "$limit_colors" = "y" ]; then
        palettegen_option="palettegen=max_colors=32"
    else
        palettegen_option="palettegen"
    fi

    output_file="${input_file%.*}.gif"

    ffmpeg -i $input_file \
        -vf "fps=${frame_rate},scale=iw*${scale_value}:ih*${scale_value}:flags=lanczos,split[s0][s1];[s0]${palettegen_option}[p];[s1][p]paletteuse" \
        -loop 0 $output_file

    echo "\n✅ Done!"
    echo "GIF size: $(du -k $output_file | awk '{print $1/1024 " MB"}')"
    echo "$output_file"
}

function 2mp4() {
    input_file=$1

    if [ -z "$input_file" ]; then
        input_file=$(ls | fzf)
    fi

    while true; do
        echo "Enter scale value (e.g., 0.75 for 75% size. Default 0.75):"
        read scale_value
        if [[ -z "$scale_value" ]]; then
            scale_value=0.75
            break
        elif [[ $scale_value =~ ^[0-9]+([.][0-9]+)?$ ]]; then
            break
        fi
    done

    while true; do
        echo "Enter frame rate (e.g. 30. Default 30):"
        read frame_rate
        if [[ -z "$frame_rate" ]]; then
            frame_rate=24
            break
        elif [[ $frame_rate =~ ^[0-9]+$ ]]; then
            break
        fi
    done

    while true; do
        echo "Enter CRF quality (lower = better/larger; 23-32 typical. Default 28):"
        read crf
        if [[ -z "$crf" ]]; then
            crf=28
            break
        elif [[ $crf =~ ^[0-9]+([.][0-9]+)?$ ]]; then
            break
        fi
    done

    output_file="${input_file%.*}.mp4"
    if [[ "$output_file" == "$input_file" ]]; then
        output_file="${input_file%.*}.h264.mp4"
    fi

    ffmpeg -i "$input_file" \
        -vf "setpts=PTS-STARTPTS,fps=${frame_rate},scale=trunc(iw*${scale_value}/2)*2:trunc(ih*${scale_value}/2)*2:flags=lanczos" \
        -r "$frame_rate" -fps_mode cfr \
        -c:v libx264 -crf "$crf" -preset slow -pix_fmt yuv420p \
        -g "$frame_rate" -keyint_min "$frame_rate" -sc_threshold 0 -bf 0 \
        -c:a aac -b:a 128k -movflags +faststart \
        "$output_file"

    echo "\n✅ Done!"
    echo "MP4 size: $(du -k "$output_file" | awk '{print $1/1024 " MB"}')"
    echo "$output_file"
}

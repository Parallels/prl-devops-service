---
layout: page
---

<div class="image toggle hand" title="Click on the image to toggle between the detailed steps and the animation">
    <div class="steps hide flex flex-column">
        <p>
            When you first start the Control Center, you will be presented with the following screen. Click on the "<span class="bold">Install Windows, Linux, or macOS from an image file</span>"button and then on "<span class="bold">Continue</span>".
        </p>
        <div class="flex flex-center">
            <img src="{{site.url}}{{site.baseurl}}/img/examples/create_vm_control_center_step1.png" alt="Create VM Control Center">
        </div>
        <p>
            Select the ISO that you just downloaded and press "<span class="bold">Continue</span>".
            If the ISO is not listed, click on the "<span class="bold">Choose Manually</span>" button and navigate to the location of the ISO file.
        </p>
        <div class="flex flex-center">
            <img src="{{site.url}}{{site.baseurl}}/img/examples/create_vm_control_center_step2.png" alt="Create VM Control Center">
        </div>
        <p>
            Give a name to the virtual machine and if you want to customize the resources select the option and press "<span class="bold">Create</span>".
        </p>
        <div class="flex flex-center">
            <img src="{{site.url}}{{site.baseurl}}/img/examples/create_vm_control_center_step3.png" alt="Create VM Control Center">
        </div>
        <p>
            If you selected the "<span class="bold">Customize settings before installation</span>" you can then press on "<span class="bold">Hardware</span>" and change the resources and other settings, like for example the network, the graphics and also the <span class="bold">Rosetta support</span>.<br/>
            Once you are done simple <span class="bold">Close</span> close the window.
        </p>
        <div class="flex flex-center">
            <div id="step4.1" class="zoom-container">
                <img src="{{site.url}}{{site.baseurl}}/img/examples/create_vm_control_center_step4.1.png" alt="Create VM Control Center">            
            </div>
            <div  id="step4.2" class="zoom-container">
                <img src="{{site.url}}{{site.baseurl}}/img/examples/create_vm_control_center_step4.2.png" alt="Create VM Control Center">
            </div>
        </div>
        <p>
            Your machine is now created and you can start it by pressing "<span class="bold">Continue</span>".
        </p>
        <div class="flex flex-center">
            <img src="{{site.url}}{{site.baseurl}}/img/examples/create_vm_control_center_step5.png" alt="Create VM Control Center">
        </div>
    </div>
    <div class="anime">
        <img src="{{site.url}}{{site.baseurl}}/img/examples/create_vm_control_center.gif" alt="Create VM Control Center">
    </div>
</div>

<script>
    document.addEventListener('DOMContentLoaded', function() {
        let div = document.querySelector('div.image.toggle');
        console.log(div);
        div.addEventListener('click', function() {
            div.querySelector('div.steps').classList.toggle('hide');
            div.querySelector('div.anime').classList.toggle('hide');
        });

        let multiColumnXZoom = document.querySelectorAll('div.zoom-container').forEach(function(el) {
            el.addEventListener('mouseenter', function() {
                el.classList.add('zoom');
                document.querySelectorAll('div.zoom-container').forEach(function(subEl) {
                    if (el.id !== subEl.id) {
                        subEl.classList.remove('zoom');
                    }
                });
            });
            el.addEventListener('mouseleave', function() {
                document.querySelectorAll('div.zoom-container').forEach(function(subEl) {
                    subEl.classList.remove('zoom');
                });
            });

        });
    });

    
</script>

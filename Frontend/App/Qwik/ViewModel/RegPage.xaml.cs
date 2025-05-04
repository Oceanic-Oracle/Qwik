using Qwik.Model.Dto;
using Qwik.Model;

namespace Qwik;

public partial class RegPage : ContentPage
{
	public RegPage()
	{
		InitializeComponent();
	}

	private async void RegisterButtonClicked(object sender, EventArgs e)
	{
        try
        {
            string login = LoginReg.Text;
            string email = EmailReg.Text;
            string password = PasswordReg.Text;
            string repPassword = PasswordRepReg.Text;

            if (string.IsNullOrEmpty(login) || string.IsNullOrEmpty(email) ||
                string.IsNullOrEmpty(password) || string.IsNullOrEmpty(repPassword))
            {
                await DisplayAlert("Ошибка", "Заполните все поля", "OK");
                return;
            }

            if (password != repPassword)
            {
                await DisplayAlert("Ошибка", "Пароли должны совпадать", "OK");
                return;
            }

            VerifyReq req = new VerifyReq { email = EmailReg.Text };

            VerifyRes res = await Api.SendPostRequest<VerifyReq, VerifyRes>(req, Api.VerifyCodeEndpoint);

            await Shell.Current.Navigation.PushAsync(new VerifyEmailPage(login, email, password, res.sessioncode));
        }
        catch (Exception ex) 
        {
            await DisplayAlert("Ошибка", $"{ex.Message}", "OK");
        }
    }
}